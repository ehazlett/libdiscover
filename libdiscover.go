package libdiscover

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/memberlist"
	"github.com/hashicorp/serf/serf"
	"github.com/sirupsen/logrus"
)

type Discover struct {
	name             string
	bindAddr         string
	advertiseAddr    string
	joinAddr         string
	cluster          *serf.Serf
	logger           *log.Logger
	userEventHandler func(e Event) error
	nodeTimeout      time.Duration
	debug            bool
}

func NewDiscover(cfg *Config) (*Discover, error) {
	d := &Discover{
		name:             cfg.Name,
		bindAddr:         cfg.BindAddr,
		advertiseAddr:    cfg.AdvertiseAddr,
		joinAddr:         cfg.JoinAddr,
		logger:           cfg.Logger,
		userEventHandler: cfg.EventHandler,
		nodeTimeout:      cfg.NodeTimeout,
		debug:            cfg.Debug,
	}

	return d, nil
}

func (d *Discover) Name() string {
	return d.name
}

func (d *Discover) LocalNode() *memberlist.Node {
	return d.cluster.Memberlist().LocalNode()
}

func (d *Discover) Members() []serf.Member {
	return d.cluster.Members()
}

func (d *Discover) Addr() string {
	return d.advertiseAddr
}

func (d *Discover) Run() error {
	mCfg := memberlist.DefaultLANConfig()
	mCfg.Logger = d.logger

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
	cfg.TombstoneTimeout = d.nodeTimeout

	// handle events
	eventChan := make(chan serf.Event)
	cfg.EventCh = eventChan

	errorChan := make(chan error)
	go func() {
		err := <-errorChan
		logrus.Error(err)
	}()

	go d.eventHandler(eventChan, errorChan)

	// set log output
	if !d.debug {
		cfg.LogOutput = ioutil.Discard
		mCfg.LogOutput = ioutil.Discard
	}

	cfg.MemberlistConfig = mCfg

	srv, err := serf.Create(cfg)
	if err != nil {
		return err
	}

	d.cluster = srv

	if d.joinAddr != "" {
		logrus.Debugf("joining cluster: addr=%s", d.joinAddr)

		if _, err := srv.Join([]string{d.joinAddr}, true); err != nil {
			return err
		}
	}

	// broadcast join event
	info := map[string]string{
		"name": d.Name(),
		"addr": d.Addr(),
	}
	data, err := json.Marshal(info)
	if err != nil {
		return err
	}
	if err := d.SendEvent("node-join", data, false); err != nil {
		return err
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
	info := map[string]string{
		"name": d.Name(),
		"addr": d.Addr(),
	}
	data, err := json.Marshal(info)
	if err != nil {
		return err
	}
	if err := d.SendEvent("node-leave", data, false); err != nil {
		return err
	}

	// leave serf cluster
	if err := d.cluster.Leave(); err != nil {
		return err
	}

	// shutdown background listeners
	if err := d.cluster.Shutdown(); err != nil {
		return err
	}

	return nil
}
