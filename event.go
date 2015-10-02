package libdiscover

import (
	log "github.com/Sirupsen/logrus"
	"github.com/hashicorp/serf/serf"
)

// eventHandler handles all events sent through the cluster
func (d *Discover) eventHandler(eventCh chan serf.Event, errCh chan error) {
	for {
		e := <-eventCh
		if err := d.handleEvent(e); err != nil {
			errCh <- err
		}
	}
}

func (d *Discover) handleEvent(evt serf.Event) error {
	log.Debugf("cluster event: %s", evt.String())

	switch evt.EventType() {
	case serf.EventMemberLeave:
		if err := d.handleMemberLeave(evt); err != nil {
			return err
		}
	case serf.EventMemberFailed:
		if err := d.handleMemberFail(evt); err != nil {
			return err
		}
	case serf.EventMemberJoin:
	case serf.EventUser:
		if err := d.handleUserEvent(evt.(serf.UserEvent)); err != nil {
			return err
		}
	}

	return nil
}

// handleUserEvent handles all user events sent through the cluster
func (d *Discover) handleUserEvent(evt serf.UserEvent) error {
	log.Debugf("user event: %s", evt.Name)

	// internal user events
	switch evt.Name {
	case "node-join":
		// only add if there is an existing cluster
		if len(d.cluster.Members()) > 1 && d.raft != nil {
			addr := string(evt.Payload)
			if err := d.addRaftPeer(addr); err != nil {
				return err
			}
		}
	case "node-leave":
		addr := string(evt.Payload)
		// remove from raft
		if err := d.removeRaftPeer(addr); err != nil {
			return err
		}
	}

	// run the custom user event handler
	if h, ok := d.handlers[evt.Name]; ok {
		go func() {
			if err := h(evt); err != nil {
				d.handlerErrCh <- err
			}
		}()
	}

	return nil
}

func (d *Discover) handleMemberLeave(evt serf.Event) error {
	if e, ok := evt.(serf.MemberEvent); ok {
		for _, m := range e.Members {
			log.Debugf("member leave: %s", m.Name)
		}
	}

	return nil
}

func (d *Discover) handleMemberFail(evt serf.Event) error {
	if e, ok := evt.(serf.MemberEvent); ok {
		for _, m := range e.Members {
			log.Debugf("member fail: %s", m.Name)
		}
	}

	return nil
}
