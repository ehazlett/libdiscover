package main

import (
	"encoding/json"
	"flag"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/ehazlett/libdiscover"
	"github.com/sirupsen/logrus"
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
	logrus.Infof("user event: name=%s data=%s", e.Name, data)

	return nil
}

func main() {
	flag.Parse()

	if flNodeName == "" {
		flNodeName = randomName()
	}

	if flDebug {
		logrus.SetLevel(logrus.DebugLevel)
		logrus.Debug("debug enabled")
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
		logrus.Fatal(err)
	}

	logrus.Infof("node id: %s", flNodeName)

	ticker := time.NewTicker(time.Millisecond * 5000)
	go func() {
		for range ticker.C {
			nodes := []string{}
			members := d.Members()

			for _, m := range members {
				nodes = append(nodes, m.Name)
			}

			logrus.Debugf("members: num=%d nodes=%s", len(members), strings.Join(nodes, ","))
			info := map[string]string{
				"name": flNodeName,
			}

			data, err := json.Marshal(info)
			if err != nil {
				logrus.Fatal(err)
			}
			if err := d.SendEvent("heartbeat", data, false); err != nil {
				logrus.Error(err)
			}
		}
	}()

	// make sure to handle this error
	if err := d.Run(); err != nil {
		logrus.Fatal(err)
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

	logrus.Debug("stopping")
	if err := d.Stop(); err != nil {
		logrus.Fatal(err)
	}
}
