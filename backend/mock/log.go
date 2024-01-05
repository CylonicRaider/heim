package mock

import (
	"sync"
	"time"

	"euphoria.leet.nu/heim/proto"
	"euphoria.leet.nu/heim/proto/snowflake"
	"euphoria.io/scope"
)

type memLog struct {
	sync.Mutex
	msgs []*proto.Message
}

func newMemLog() *memLog { return &memLog{msgs: []*proto.Message{}} }

func (log *memLog) post(msg *proto.Message) {
	log.Lock()
	defer log.Unlock()

	log.msgs = append(log.msgs, msg)
}

func (log *memLog) GetMessage(ctx scope.Context, id snowflake.Snowflake) (*proto.Message, error) {
	log.Lock()
	defer log.Unlock()

	for _, msg := range log.msgs {
		if msg.ID == id {
			return msg, nil
		}
	}
	return nil, proto.ErrMessageNotFound
}

func (log *memLog) Latest(ctx scope.Context, n int, before snowflake.Snowflake) ([]proto.Message, error) {
	log.Lock()
	defer log.Unlock()

	end := len(log.msgs)
	if !before.IsZero() {
		for end > 0 && !log.msgs[end-1].ID.Before(before) {
			end--
		}
	}

	start := end - n
	if start < 0 {
		start = 0
	}

	slice := make([]*proto.Message, 0, n)
	for _, msg := range log.msgs[start:] {
		if time.Time(msg.Deleted).IsZero() {
			slice = append(slice, maybeTruncate(msg))
			if len(slice) >= n {
				break
			}
		}
	}
	if len(slice) == 0 {
		return []proto.Message{}, nil
	}

	messages := make([]proto.Message, len(slice))
	for i, msg := range slice {
		messages[i] = *msg
	}
	return messages, nil
}

func (log *memLog) edit(e proto.EditMessageCommand) (*proto.Message, error) {
	log.Lock()
	defer log.Unlock()

	now := proto.Now()
	for _, msg := range log.msgs {
		if msg.ID == e.ID {
			if e.Parent != 0 {
				msg.Parent = e.Parent
			}
			if e.Content != "" {
				msg.Content = e.Content
			}
			if e.Delete {
				msg.Deleted = now
			} else {
				msg.Deleted = proto.Time{}
			}
			msg.Edited = now
			return maybeTruncate(msg), nil
		}
	}
	return nil, proto.ErrMessageNotFound
}

func maybeTruncate(msg *proto.Message) *proto.Message {
	if len(msg.Content) > proto.MaxMessageTransmissionLength {
		truncated := *msg
		if msg.EncryptionKeyID != "" {
			truncated.Content = ""
		} else {
			truncated.Content = truncated.Content[:proto.MaxMessageTransmissionLength]
		}
		truncated.Truncated = true
		msg = &truncated
	}
	return msg
}
