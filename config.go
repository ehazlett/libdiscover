package libdiscover

import (
	"net"

	"github.com/hashicorp/raft"
	"github.com/hashicorp/serf/serf"
)

type Config struct {
	Name              string
	BindAddr          string
	BindPort          int
	AdvertiseAddr     string
	AdvertisePort     int
	RaftBindAddr      string
	RaftAdvertiseAddr net.Addr
	JoinAddr          string
	StorePath         string
	Handlers          map[string]func(e serf.UserEvent) error
	HandlerErrCh      chan error
	FSM               raft.FSM
	Debug             bool
}
