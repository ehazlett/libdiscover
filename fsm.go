package libdiscover

import (
	"io"

	log "github.com/Sirupsen/logrus"
	"github.com/hashicorp/raft"
)

// GridFSM is a sample implementation of the finate state machine
type GridFSM struct{}

func (g GridFSM) Apply(l *raft.Log) interface{} {
	log.Debugf("fsm: apply: data=%s", string(l.Data))
	return nil
}

func (g GridFSM) Snapshot() (raft.FSMSnapshot, error) {
	log.Debug("fsm: snapshot")
	return GridFSMSnapshot{}, nil
}

func (g GridFSM) Restore(r io.ReadCloser) error {
	log.Debug("fsm: restore")
	return nil
}
