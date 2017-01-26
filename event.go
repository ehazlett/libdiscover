package libdiscover

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/hashicorp/serf/serf"
	"github.com/sirupsen/logrus"
)

type Event struct {
	serf.UserEvent
	Created int64 `json:"created"`
	Data    interface{}
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
		se := evt.(serf.UserEvent)

		var data interface{}
		if err := json.Unmarshal(se.Payload, &data); err != nil {
			logrus.Errorf("payload: %v", string(se.Payload))
			return fmt.Errorf("error unmarshalling payload: %s", err)
		}

		e := Event{
			se,
			time.Now().Unix(),
			data,
		}

		if err := d.userEventHandler(e); err != nil {
			return err
		}
	}

	return nil
}

func (d *Discover) handleMemberLeave(evt serf.Event) error {
	if e, ok := evt.(serf.MemberEvent); ok {
		for _, m := range e.Members {
			logrus.Debugf("member leave: %s", m.Name)
		}
	}

	return nil
}

func (d *Discover) handleMemberFail(evt serf.Event) error {
	if e, ok := evt.(serf.MemberEvent); ok {
		for _, m := range e.Members {
			logrus.Debugf("member fail: %s", m.Name)
		}
	}

	return nil
}
