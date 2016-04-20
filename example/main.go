package main

import (
	"flag"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/ehazlett/libdiscover"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

var (
	flNodeName      string
	flBindAddr      string
	flAdvertiseAddr string
	flJoinAddr      string
	flNodeTimeout   int
	flDebug         bool
	flClusterDebug  bool
)

type heartbeatInfo struct {
	Name string `json:"name"`
}

func init() {
	flag.StringVar(&flNodeName, "name", "", "bind address")
	flag.StringVar(&flBindAddr, "bind", "127.0.0.1:7946", "bind address")
	flag.StringVar(&flAdvertiseAddr, "advertise", "127.0.0.1:7946", "advertise address")
	flag.IntVar(&flNodeTimeout, "timeout", 60, "node timeout (seconds)")
	flag.StringVar(&flJoinAddr, "join", "", "join address")
	flag.BoolVar(&flDebug, "debug", false, "enable debug")
	flag.BoolVar(&flClusterDebug, "cluster-debug", false, "enable cluster debug messages")
}

func randomName() string {
	b := make([]byte, 8)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}

	return strings.ToLower("node-" + string(b))
}

func eventHandler(e libdiscover.Event) error {
	data := string(e.Payload)
	log.Infof("user event: name=%s data=%s", e.Name, data)

	return nil
}

func main() {
	flag.Parse()

	if flNodeName == "" {
		flNodeName = randomName()
	}

	if flDebug {
		log.SetLevel(log.DebugLevel)
		log.Debug("debug enabled")
	}

	if flNodeTimeout == 0 {
		flNodeTimeout = 60
	}

	cfg := &libdiscover.Config{
		Name:          flNodeName,
		BindAddr:      flBindAddr,
		AdvertiseAddr: flAdvertiseAddr,
		JoinAddr:      flJoinAddr,
		Debug:         flClusterDebug,
		NodeTimeout:   time.Second * time.Duration(flNodeTimeout),
		EventHandler:  eventHandler,
	}

	// make sure to handle this error
	d, err := libdiscover.NewDiscover(cfg)
	if err != nil {
		log.Fatal(err)
	}

	log.Infof("node id: %s", flNodeName)

	ticker := time.NewTicker(time.Millisecond * 5000)
	go func() {
		for range ticker.C {
			nodes := []string{}
			members := d.Members()

			for _, m := range members {
				nodes = append(nodes, m.Name)
			}

			log.Debugf("members: num=%d nodes=%s", len(members), strings.Join(nodes, ","))
			if err := d.SendEvent("heartbeat", map[string]interface{}{
				"name": flNodeName,
			}, false); err != nil {
				log.Error(err)
			}
		}
	}()

	// make sure to handle this error
	if err := d.Run(); err != nil {
		log.Fatal(err)
	}

	// handle interrupt
	signals := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-signals
		done <- true
	}()

	<-done

	log.Debug("stopping")
	if err := d.Stop(); err != nil {
		log.Fatal(err)
	}
}
