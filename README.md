# libdiscover
libdiscover is a library for building distributed systems.  It uses Gossip via 
[Serf](https://serfdom.io/) for node discovery and 
[Raft](https://github.com/hashicorp/raft) for replicated log and state
machines.

# Example

```go
cfg := &libdiscover.Config{
    Name: "node-00",
    BindAddr: "127.0.0.1:7946",
    AdvertiseAddr: "127.0.0.1:7946",
    RaftBindAddr: "127.0.0.1:8946",
    RaftAdvertiseAddr: "127.0.0.1:8946",
    JoinAddr: "",
    StorePath: "/tmp/store",
    Debug: true,
}

// make sure to handle this error
d, _ := libdiscover.NewDiscover(cfg, nil)

// make sure to handle this error
_ = d.Run()
```

This will start everything needed to start discovery and replication.  To
join another node, simply update the network bind and advertise settings
and set the `JoinAddr` to an address of any peer (i.e. `127.0.0.1:7946`).

# FSM
The included FSM should only be used for debug.  You will most likely want to
bring your own.  See the example for details on how to implement it.
You can pass in your custom FSM like so:

```go
fsm := &MyFSM{}

d, _ := libdiscover.NewDiscover(cfg, fsm)
```

# Handlers
libdiscover has the ability to broadcast custom user events.  To handle custom
events, you can either pass a `map[string]func(e serf.UserEvent) error`
upon initialization or via the `SetHandlers` func.  Here is an example:

```go
eventHandlers := map[string]func(e serf.UserEvent) error{
	"node-join":    nodeJoinHandler,
}

eventErrCh := make(chan error)

d.SetHandlers(eventHandlers, eventErrCh)
```

This would be the event handler.

```go
func nodeJoinHandler(e serf.UserEvent) error {
    log.Println(string(e.Payload))

    return nil
}
```
