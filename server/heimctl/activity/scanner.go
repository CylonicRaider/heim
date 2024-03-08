package activity

import (
	"encoding/json"

	"github.com/euphoria-io/scope"
	"github.com/lib/pq"

	"euphoria.leet.nu/heim/backend/psql"
	"euphoria.leet.nu/heim/proto"
	"euphoria.leet.nu/heim/proto/logging"
)

func ScanLoop(ctx scope.Context, listener *pq.Listener) {
	defer ctx.WaitGroup().Done()

	logger := logging.Logger(ctx)
	for {
		select {
		case <-ctx.Done():
			logger.Printf("received cancellation signal, shutting down")
			return
		case notice := <-listener.Notify:
			if notice == nil {
				logger.Printf("received nil from listener")
				continue
			}

			var msg psql.BroadcastMessage

			if err := json.Unmarshal([]byte(notice.Extra), &msg); err != nil {
				logger.Printf("error: pq listen: invalid broadcast: %s", err)
				logger.Printf("         payload: %#v", notice.Extra)
				continue
			}

			switch msg.Event.Type {
			case proto.BounceEventType:
				bounceActivity.WithLabelValues(msg.Room).Inc()
			case proto.JoinEventType:
				joinActivity.WithLabelValues(msg.Room).Inc()
			case proto.PartEventType:
				partActivity.WithLabelValues(msg.Room).Inc()
			case proto.SendEventType:
				messageActivity.WithLabelValues(msg.Room).Inc()
			}
		}
	}
}
