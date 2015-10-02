package libdiscover

import (
	"github.com/hashicorp/raft"
)

type GridFSMSnapshot struct{}

func (s GridFSMSnapshot) Persist(sink raft.SnapshotSink) error {
	return nil
}

func (s GridFSMSnapshot) Release() {}
