package libdiscover

import (
	"log"
	"time"
)

type Config struct {
	Name          string
	BindAddr      string
	AdvertiseAddr string
	JoinAddr      string
	Logger        *log.Logger
	EventHandler  func(e Event) error
	NodeTimeout   time.Duration
	Debug         bool
}
