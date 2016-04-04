# libdiscover
libdiscover is a library for building distributed systems.  It uses Gossip via 
[Serf](https://serfdom.io/) for node discovery.

# Example

```go
cfg := &libdiscover.Config{
    Name: "node-00",
    BindAddr: "127.0.0.1:7946",
    AdvertiseAddr: "127.0.0.1:7946",
    JoinAddr: "",
    NodeTimeout:   time.Millisecond * 10000,
    Debug: true,
}

// make sure to handle this error
d, _ := libdiscover.NewDiscover(cfg)

// make sure to handle this error
_ = d.Run()
```

This will start everything needed to start discovery and replication.  To
join another node, simply update the network bind and advertise settings
and set the `JoinAddr` to an address of any peer (i.e. `127.0.0.1:7946`).

# Event Handler
libdiscover has the ability to broadcast custom user events.  To handle custom
events, you can pass a `func(e serf.UserEvent) error` in the
libdiscover.Config.

```go
func eventHandler(e serf.UserEvent) error {
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
```
