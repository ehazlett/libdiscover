package libdiscover

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/hashicorp/raft"
	boltdb "github.com/hashicorp/raft-boltdb"
)

func (d *Discover) initRaft(enableSingleNode bool) error {
	logsPath := filepath.Join(d.storePath, "logs")
	stablePath := filepath.Join(d.storePath, "raft.db")
	snapsPath := filepath.Join(d.storePath, "snaps")
	peersPath := filepath.Join(d.storePath)

	if _, err := os.Stat(d.storePath); err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(d.storePath, 0700); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	logs, err := boltdb.NewBoltStore(logsPath)
	if err != nil {
		return err
	}

	d.raftStore, err = boltdb.NewBoltStore(stablePath)
	if err != nil {
		return err
	}
	stableStore := d.raftStore

	snapshotStore, err := raft.NewFileSnapshotStore(snapsPath, 1, nil)
	if err != nil {
		return err
	}

	// raft config
	defaultCfg := raft.DefaultConfig()
	defaultCfg.EnableSingleNode = enableSingleNode
	defaultCfg.DisableBootstrapAfterElect = true
	defaultCfg.ShutdownOnRemove = false

	// default fsm if nil
	if d.fsm == nil {
		d.fsm = &GridFSM{}
	}

	log.Debugf("raft transport: bind=%s advertise=%s", d.raftBindAddr, d.raftAdvertiseAddr.String())
	transport, err := raft.NewTCPTransport(
		d.raftBindAddr,
		d.raftAdvertiseAddr,
		10,
		time.Second*5,
		nil,
	)
	if err != nil {
		return err
	}

	d.raftPeerStore = raft.NewJSONPeers(peersPath, transport)
	peerStore := d.raftPeerStore

	r, err := raft.NewRaft(
		defaultCfg,
		d.fsm,
		logs,
		stableStore,
		snapshotStore,
		peerStore,
		transport,
	)
	if err != nil {
		return err
	}

	d.raft = r

	// debug
	if d.debug {
		ticker := time.NewTicker(time.Second * 10)
		go func() {
			for _ = range ticker.C {
				log.Debugf("raft state: %s", d.raft.State().String())

				peers, err := d.raftPeerStore.Peers()
				if err != nil {
					log.Error(err)
					continue
				}
				log.Debugf("raft peers: %s", strings.Join(peers, ", "))

				log.Debugf("raft stats: %v", d.raft.Stats())
			}
		}()
	}

	return nil
}

func (d *Discover) addRaftPeer(addr string) error {
	if d.raft.State() == raft.Leader {
		log.Debugf("adding peer: addr=%s", addr)

		if future := d.raft.AddPeer(addr); future.Error() != nil {
			return future.Error()
		}
	}

	return nil
}

func (d *Discover) removeRaftPeer(addr string) error {
	// remove from raft
	if d.raft.State() == raft.Leader {
		log.Debugf("removing peer: addr=%s", addr)
		if future := d.raft.RemovePeer(addr); future.Error() != nil {
			return future.Error()
		}
	}

	return nil
}
