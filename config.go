package libdiscover

import (
	"log"
	"time"

	"github.com/hashicorp/serf/serf"
)

type Config struct {
	Name          string
	BindAddr      string
	AdvertiseAddr string
	JoinAddr      string
	Logger        *log.Logger
	EventHandler  func(e serf.UserEvent) error
	NodeTimeout   time.Duration
	Debug         bool
}
