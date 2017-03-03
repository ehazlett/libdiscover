# libdiscover
libdiscover is a library for building distributed systems.  It uses Gossip via
[Serf](https://serfdom.io/) for node discovery.

# Example
libdiscover has the ability to broadcast custom user events.  To handle
events, you can pass a `func(e libdiscover.Event) error` in the
libdiscover.Config.

```go
func eventHandler(e libdiscover.Event) error {
    fmt.Println(e)

    return nil
}

cfg := &libdiscover.Config{
    Name: "node-00",
    BindAddr: "127.0.0.1:7946",
    AdvertiseAddr: "127.0.0.1:7946",
    JoinAddr: "",
    NodeTimeout:   time.Millisecond * 10000,
    EventHandler: eventHandler,
    Debug: true,
}

// create the discovery server; be sure to handle the erro
d, _ := libdiscover.NewDiscover(cfg)

// start the node; be sure to handle the error
_ = d.Run()

// send an event ; be sure to handle the error
_ = d.SendEvent("test-event", []byte("testing"), false)
```

This will start everything needed to start discovery and replication.  To
join another node, simply update the network bind and advertise settings
and set the `JoinAddr` to an address of any peer (i.e. `127.0.0.1:7946`).
