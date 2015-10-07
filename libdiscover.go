package libdiscover

import (
	"net"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/hashicorp/memberlist"
	"github.com/hashicorp/raft"
	boltdb "github.com/hashicorp/raft-boltdb"
	"github.com/hashicorp/serf/serf"
)

type Discover struct {
	name              string
	bindAddr          string
	bindPort          int
	advertiseAddr     string
	advertisePort     int
	raftBindAddr      string
	raftAdvertiseAddr net.Addr
	joinAddr          string
	storePath         string
	handlers          map[string]func(e serf.UserEvent) error
	handlerErrCh      chan error
	fsm               raft.FSM
	raft              *raft.Raft
	raftStore         *boltdb.BoltStore
	raftPeerStore     raft.PeerStore
	cluster           *serf.Serf
	debug             bool
}

func NewDiscover(cfg *Config, fsm raft.FSM) (*Discover, error) {
	raftAdvAddr, err := net.ResolveTCPAddr("tcp", cfg.RaftAdvertiseAddr)
	if err != nil {
		return nil, err
	}

	s := &Discover{
		name:              cfg.Name,
		bindAddr:          cfg.BindAddr,
		advertiseAddr:     cfg.AdvertiseAddr,
		raftBindAddr:      cfg.RaftBindAddr,
		raftAdvertiseAddr: raftAdvAddr,
		joinAddr:          cfg.JoinAddr,
		storePath:         cfg.StorePath,
		handlers:          cfg.Handlers,
		handlerErrCh:      cfg.HandlerErrCh,
		fsm:               fsm,
		debug:             cfg.Debug,
	}

	return s, nil
}

func (d *Discover) SetHandlers(h map[string]func(e serf.UserEvent) error, c chan error) {
	d.handlers = h
	d.handlerErrCh = c
}

func (d *Discover) LocalNode() *memberlist.Node {
	return d.cluster.Memberlist().LocalNode()
}

func (d *Discover) Members() []serf.Member {
	return d.cluster.Members()
}

func (d *Discover) Run() error {
	mCfg := memberlist.DefaultLANConfig()

	bindAddr := "127.0.0.1"
	bindPort := 7946

	bindParts := strings.Split(d.bindAddr, ":")

	if len(bindParts) > 1 {
		bindAddr = bindParts[0]
		p, err := strconv.Atoi(bindParts[1])
		if err != nil {
			return err
		}

		bindPort = p
	}

	advertiseAddr := "127.0.0.1"
	advertisePort := 7946

	advParts := strings.Split(d.bindAddr, ":")

	if len(advParts) > 1 {
		advertiseAddr = advParts[0]
		p, err := strconv.Atoi(advParts[1])
		if err != nil {
			return err
		}

		advertisePort = p
	}

	mCfg.Name = d.name
	mCfg.BindAddr = bindAddr
	mCfg.BindPort = bindPort
	mCfg.AdvertiseAddr = advertiseAddr
	mCfg.AdvertisePort = advertisePort

	cfg := serf.DefaultConfig()
	cfg.NodeName = d.name
	cfg.MemberlistConfig = mCfg

	// handle events
	eventChan := make(chan serf.Event)
	cfg.EventCh = eventChan

	errorChan := make(chan error)
	go func() {
		err := <-errorChan
		log.Error(err)
	}()

	go d.eventHandler(eventChan, errorChan)

	srv, err := serf.Create(cfg)
	if err != nil {
		return err
	}

	d.cluster = srv

	enableSingleNode := true
	if d.joinAddr != "" {
		log.Debugf("joining cluster: addr=%s", d.joinAddr)

		if _, err := srv.Join([]string{d.joinAddr}, true); err != nil {
			return err
		}

		enableSingleNode = false
	}

	// broadcast join event
	if err := d.cluster.UserEvent("node-join", []byte(d.raftAdvertiseAddr.String()), false); err != nil {
		return err
	}

	if err := d.initRaft(enableSingleNode); err != nil {
		return err
	}

	return nil
}

// IsLeader returns a boolean whether the node is the cluster leader
func (d *Discover) IsLeader() bool {
	if d.raft.State() == raft.Leader {
		return true
	}

	return false
}

// Apply issues a command to the FSM
func (d *Discover) Apply(cmd []byte, timeout time.Duration) error {
	if d.raft.State() == raft.Leader {
		if f := d.raft.Apply(cmd, timeout); f.Error() != nil {
			return f.Error()
		}
	}

	return nil
}

// SendEvent allows for sending custom events in the cluster
func (d *Discover) SendEvent(name string, data []byte, coalesce bool) error {
	if err := d.cluster.UserEvent(name, data, coalesce); err != nil {
		return err
	}

	return nil
}

// Stop shuts down the cluster
func (d *Discover) Stop() error {
	// broadcast node leave
	if err := d.cluster.UserEvent("node-leave", []byte(d.raftAdvertiseAddr.String()), false); err != nil {
		return err
	}

	// wait for replication
	time.Sleep(time.Millisecond * 100)

	// shutdown raft
	if future := d.raft.Shutdown(); future.Error() != nil {
		return future.Error()
	}

	// leave serf cluster
	if err := d.cluster.Leave(); err != nil {
		return err
	}

	return nil
}

// StorePath returns the base store path for the cluster
func (d *Discover) StorePath() string {
	return d.storePath
}
