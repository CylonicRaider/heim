/* Rewrite of github.com/sdming/gosnow */

package snowflake

import (
	"crypto/rand"
	"fmt"
	"hash/crc32"
	"net"
	"sync"
	"time"
)

const (
	WorkerIDBits = 10
	MaxWorkerID  = (1 << WorkerIDBits) - 1
	SequenceBits = 12
	MaxSequence  = (1 << SequenceBits) - 1
)

var (
	DefaultWorkerID = defaultWorkerID()
)

func timestampAt(now time.Time) uint64 {
	return uint64(time.Duration(now.UnixNano()) / time.Millisecond)
}

func timestamp() uint64 {
	return timestampAt(time.Now())
}

func assemble(timestamp uint64, workerID uint16, sequence uint16) uint64 {
	return timestamp<<(WorkerIDBits+SequenceBits) |
		uint64(workerID<<SequenceBits) |
		uint64(sequence)
}

type SnowflakeGen struct {
	sync.Mutex
	workerID  uint16
	sequence  uint16
	epoch     uint64
	timestamp uint64
}

func NewSnowflakeGen(epoch time.Time, workerID uint16) (*SnowflakeGen, error) {
	if workerID > MaxWorkerID {
		return nil, fmt.Errorf("worker ID out of range")
	}
	return &SnowflakeGen{
		workerID: workerID,
		epoch:    timestampAt(epoch),
	}, nil
}

func (g *SnowflakeGen) Next() (uint64, error) {
	g.Lock()
	defer g.Unlock()

	now := time.Now()
	nowTS := timestampAt(now)
	if nowTS > g.timestamp {
		g.timestamp = nowTS
		g.sequence = 0
	} else if nowTS < g.timestamp {
		return 0, fmt.Errorf("time went backwards")
	} else if g.sequence < MaxSequence {
		g.sequence++
	} else {
		var err error
		g.timestamp, err = waitForNextMillisecond(now, nowTS)
		if err != nil {
			return 0, err
		}
		g.sequence = 0
	}

	return assemble(g.timestamp-g.epoch, g.workerID, g.sequence), nil
}

func waitForNextMillisecond(start time.Time, startTS uint64) (uint64, error) {
	// this should wake us up right at the start of the next millisecond
	time.Sleep(time.Millisecond - time.Duration(start.Nanosecond())%time.Millisecond)
	for {
		now := time.Now()
		if nowTS := timestampAt(now); nowTS > startTS {
			return nowTS, nil
		} else if nowTS < startTS {
			return 0, fmt.Errorf("time went backwards")
		}
		// surely the millisecond timestamp will be different a millisecond later?
		time.Sleep(time.Millisecond)
	}
}

func defaultWorkerID() uint16 {
	ifs, err := net.Interfaces()
	if err == nil {
		buffer := make([]byte, 2)
		rand.Read(buffer)
		return (uint16(buffer[0])<<8 | uint16(buffer[1])) & MaxWorkerID
	}
	hasher := crc32.NewIEEE()
	for _, i := range ifs {
		hasher.Write(i.HardwareAddr)
	}
	return uint16(hasher.Sum32() & MaxWorkerID)
}
