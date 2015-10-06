package libdiscover

import (
	"github.com/hashicorp/raft"
	"github.com/hashicorp/serf/serf"
)

type Config struct {
	Name              string
	BindAddr          string
	AdvertiseAddr     string
	RaftBindAddr      string
	RaftAdvertiseAddr string
	JoinAddr          string
	StorePath         string
	Handlers          map[string]func(e serf.UserEvent) error
	HandlerErrCh      chan error
	FSM               raft.FSM
	Debug             bool
}
