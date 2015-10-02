# libdiscover
Libdiscover is a library for distributed systems.  It uses Gossip via 
[Serf](https://serfdom.io/) for node discovery and 
[Raft](https://github.com/hashicorp/raft) for replicated log and state
machines.

# Example

```go
cfg := &libdiscover.DiscoverConfig{
    Name: "node-00",
    BindAddr: "127.0.0.1",
    BindPort: 7946,
    AdvertiseAddr: "127.0.0.1",
    AdvertisePort: 7946,
    RaftBindAddr: "127.0.0.1",
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
