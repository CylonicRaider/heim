package snowflake

import (
	"encoding/json"
	"fmt"
	"strconv"
	"sync/atomic"
	"time"
)

type Snowflaker interface {
	Next() (uint64, error)
}

func Clock() time.Time { return time.Now() }

var Epoch = time.Date(2014, 12, 0, 0, 0, 0, 0, time.UTC)

var DefaultSnowflaker = func() Snowflaker {
	result, err := NewSnowflakeGen(Epoch, DefaultWorkerID)
	if err != nil {
		panic(err)
	}
	return result
}()

type Snowflake uint64

func New() (Snowflake, error) {
	snowflake, err := DefaultSnowflaker.Next()
	if err != nil {
		return Snowflake(0), err
	}
	return Snowflake(snowflake), nil
}

func NewFromTime(t time.Time, counter *uint64) Snowflake {
	seqID := atomic.AddUint64(counter, 1)
	return Snowflake(assemble(
		timestampAt(t)-timestampAt(Epoch),
		DefaultWorkerID,
		uint16(seqID),
	))
}

func NewFromString(s string) (Snowflake, error) {
	snowflake, err := strconv.ParseUint(s, 36, 64)
	if err != nil {
		return Snowflake(0), err
	}
	return Snowflake(snowflake), nil
}

func (s Snowflake) String() string {
	if s == 0 {
		return ""
	}
	return fmt.Sprintf("%013s", strconv.FormatUint(uint64(s), 36))
}

func (s Snowflake) GoString() string { return fmt.Sprintf("%v", s.String()) }

func (s *Snowflake) FromString(str string) error {
	if str == "" {
		*s = 0
		return nil
	}

	i, err := strconv.ParseUint(str, 36, 64)
	if err != nil {
		return err
	}

	*s = Snowflake(i)
	return nil
}

func (s Snowflake) MarshalJSON() ([]byte, error) { return json.Marshal(s.String()) }

func (s *Snowflake) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	return s.FromString(str)
}

func (s Snowflake) Time() time.Time {
	timestampMillis := uint64(s) >> (WorkerIDBits + SequenceBits)
	return Epoch.Add(time.Duration(timestampMillis) * time.Millisecond)
}

func (s Snowflake) IsZero() bool                    { return s == 0 }
func (s Snowflake) Before(reference Snowflake) bool { return s < reference }
