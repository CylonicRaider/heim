package psql

import (
	"fmt"
	"strings"

	"euphoria.leet.nu/heim/proto"
	"euphoria.leet.nu/heim/proto/logging"
	"github.com/euphoria-io/scope"
)

type Listener struct {
	proto.Session
	*proto.Client
	enabled bool
}

type ListenerMap map[string]Listener

func (lm ListenerMap) Broadcast(ctx scope.Context, event *proto.Packet, exclude ...string) error {
	payload, err := event.Payload()
	if err != nil {
		return err
	}

	excludeSet := map[string]struct{}{}
	for _, exc := range exclude {
		excludeSet[exc] = struct{}{}
	}

	// Inspect packet to see if it's a bounce event. If so, we'll deliver it
	// only to the bounced parties.
	bounceAgentID := proto.UserID("")
	bounceIP := ""
	if event.Type == proto.BounceEventType {
		if bounceEvent, ok := payload.(*proto.BounceEvent); ok {
			bounceAgentID = bounceEvent.AgentID
			bounceIP = bounceEvent.IP
		} else {
			logging.Logger(ctx).Printf("wtf? expected *proto.BounceEvent, got %T", payload)
		}
	}

	// Inspect packet to see if it's a join event. If so, we'll enable the excluded
	// listener, and look for aliased sessions to kick into fast-keepalive mode.
	fastKeepaliveAgentID := ""
	if event.Type == proto.JoinEventType {
		if presence, ok := payload.(*proto.PresenceEvent); ok {
			if idx := strings.IndexRune(string(presence.ID), '-'); idx >= 1 {
				fastKeepaliveAgentID = string(presence.ID[:idx])
			}
		}
		for _, sessionID := range exclude {
			listener, ok := lm[sessionID]
			if ok && !listener.enabled {
				listener.enabled = true
				lm[sessionID] = listener
			}
		}
	}

	for sessionID, listener := range lm {
		if _, ok := excludeSet[sessionID]; !ok {
			if bounceAgentID != "" {
				if listener.Session.Identity().ID() == bounceAgentID {
					logging.Logger(ctx).Printf("sending disconnect to %s: %#v", listener.ID(), payload)
					discEvent := &proto.DisconnectEvent{Reason: payload.(*proto.BounceEvent).Reason}
					if err := listener.Send(ctx, proto.DisconnectEventType, discEvent); err != nil {
						logging.Logger(ctx).Printf("error sending disconnect event to %s: %s",
							listener.ID(), err)
					}
				}
				continue
			}
			if bounceIP != "" {
				if listener.Client.IP == bounceIP {
					logging.Logger(ctx).Printf("sending disconnect to %s: %#v", listener.ID(), payload)
					discEvent := &proto.DisconnectEvent{Reason: payload.(*proto.BounceEvent).Reason}
					if err := listener.Send(ctx, proto.DisconnectEventType, discEvent); err != nil {
						logging.Logger(ctx).Printf("error sending disconnect event to %s: %s",
							listener.ID(), err)
					}
				}
				continue
			}
			if fastKeepaliveAgentID != "" && strings.HasPrefix(sessionID, fastKeepaliveAgentID) {
				if err := listener.CheckAbandoned(); err != nil {
					fmt.Errorf("fast keepalive to %s: %s", listener.ID(), err)
				}
			}
			if !listener.enabled {
				// The event occurred before the listener joined, so don't deliver it.
				logging.Logger(ctx).Printf("not delivering event %s before %s joined",
					event.Type, listener.ID())
				continue
			}
			if err := listener.Send(ctx, event.Type, payload); err != nil {
				// TODO: accumulate errors
				return fmt.Errorf("send message to %s: %s", listener.ID(), err)
			}
		}
	}

	return nil
}

func (lm ListenerMap) NotifyUser(ctx scope.Context, userID proto.UserID, event *proto.Packet, exclude ...string) error {
	excludeSet := map[string]struct{}{}
	for _, exc := range exclude {
		excludeSet[exc] = struct{}{}
	}
	payload, err := event.Payload()
	if err != nil {
		return err
	}
	kind, id := userID.Parse()
	for sessionID, listener := range lm {
		// check that the listener is not excluded
		if _, ok := excludeSet[sessionID]; ok {
			continue
		}

		if listener.Identity().ID() == userID || (kind == "agent" && id == listener.AgentID()) {
			listener.Send(ctx, event.Type, payload)
		}
	}
	return nil
}
