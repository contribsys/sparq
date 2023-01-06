package util

/*
MIT License

Copyright (c) 2021 HotPotatoC

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

import (
	"os"
	"sync"
	"time"
)

const (
	fieldBits    = 10
	sequenceBits = 12
	maxFieldBits = 0x3FF // 0x3FF shorthand for (1 << fieldBits) - 1 or 1023
	maxSeqBits   = 0xFFF // 0xFFF shorthand for (1 << sequenceBits) - 1 or 4095
)

var (
	epoch = time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)
)

func Epoch() time.Time { return epoch }

type Snowflake struct {
	mtx         sync.Mutex
	pid         int
	sequence    uint64
	elapsedTime int64
	lastID      uint64
}

// New returns a new snowflake.ID (max field value: 1023)
func NewSnowflake() *Snowflake {
	return &Snowflake{pid: os.Getpid(), lastID: 0}
}

// NextID returns a new snowflake ID.
func (id *Snowflake) NextSID() string {
	sid := id.NextID()
	ssid := compress(sid)
	return ssid
}

var alphabet = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_")

// XXX Little endian only!
func compress(num uint64) string {
	buf := make([]rune, 11)
	buf[10] = alphabet[num&63]
	buf[9] = alphabet[(num>>6)&63]
	buf[8] = alphabet[(num>>12)&63]
	buf[7] = alphabet[(num>>18)&63]
	buf[6] = alphabet[(num>>24)&63]
	buf[5] = alphabet[(num>>30)&63]
	buf[4] = alphabet[(num>>36)&63]
	buf[3] = alphabet[(num>>42)&63]
	buf[2] = alphabet[(num>>48)&63]
	buf[1] = alphabet[(num>>54)&63]
	buf[0] = alphabet[(num>>60)&63]
	return string(buf)
}

func (id *Snowflake) NextID() uint64 {
	id.mtx.Lock()
	defer id.mtx.Unlock()

	nowSinceEpoch := msSinceEpoch()

	if nowSinceEpoch == id.elapsedTime { // same millisecond as last time
		id.sequence = (id.sequence + 1) & maxSeqBits // increment sequence number

		if id.sequence == 0 {
			// sequence overflow
			nowSinceEpoch = waitUntilNextMs(id.elapsedTime) // wait until next millisecond
		}
	} else {
		id.sequence = 0
	}

	id.elapsedTime = nowSinceEpoch

	timestampSegment := uint64(id.elapsedTime << (sequenceBits + fieldBits))
	fieldSegment := uint64(id.pid) << sequenceBits
	sequenceSegment := uint64(id.sequence)

	// if the field is bigger than the max, we need to reset it
	if id.pid > maxFieldBits {
		fieldSegment = 0
	}

	return timestampSegment | fieldSegment | sequenceSegment
}

// SID is the parsed representation of a snowflake ID.
type SID struct {
	// Timestamp is the timestamp of the snowflake ID.
	Timestamp int64
	// Sequence is the sequence number of the snowflake ID.
	Sequence uint64
	// Field is the field value of the snowflake ID.
	Field uint64
}

// Parse parses an existing snowflake ID
func Parse(sid uint64) SID {
	return SID{
		Timestamp: getTimestamp(sid),
		Sequence:  getSequence(sid),
		Field:     getDiscriminant(sid),
	}
}

// waitUntilNextMs waits until the next millisecond to return. (internal-use only)
func waitUntilNextMs(last int64) int64 {
	ms := msSinceEpoch()
	for ms <= last {
		ms = msSinceEpoch()
	}
	return ms
}

// msSinceEpoch returns the number of milliseconds since the epoch. (internal-use only)
func msSinceEpoch() int64 {
	return time.Since(epoch).Nanoseconds() / 1e6
}

// getDiscriminant returns the discriminant value of a snowflake ID. (internal-use only)
func getDiscriminant(id uint64) uint64 {
	return (id >> sequenceBits) & maxFieldBits
}

// getTimestamp returns the timestamp of a snowflake ID. (internal-use only)
func getTimestamp(id uint64) int64 {
	return int64(id>>(sequenceBits+fieldBits)) + epoch.UnixNano()/1e6
}

// getSequence returns the sequence number of a snowflake ID. (internal-use only)
func getSequence(id uint64) uint64 { return uint64(int(id) & maxSeqBits) }
