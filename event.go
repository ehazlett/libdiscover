package libdiscover

import (
	log "github.com/Sirupsen/logrus"
	"github.com/hashicorp/serf/serf"
)

type Event struct {
	serf.UserEvent
}

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
		if err := d.userEventHandler(evt.(Event)); err != nil {
			return err
		}
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
