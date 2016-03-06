package snowflake

import (
	"strconv"
	"sync"
	"time"
)

const nanosInMilli int64 = 1000 * 1000

var epoch int64 = time.Date(2016, time.January, 1, 0, 0, 0, 0, time.UTC).UnixNano() / nanosInMilli

const (
	ServerBits   = 10
	SequenceBits = 12
)

// Default represents a default Factory suitable for unit testing; this is NOT suitable for production
// use as the ServerId is fixed at 0
var Default = New(Options{
	ServerBits:   10,
	SequenceBits: 12,
})

// Options contains the configurable options for Factory
type Options struct {
	// ServerId represents a unique value that identifies this generator instance
	ServerId int64

	// ServerBits represents the number of bits used to represents the server
	ServerBits uint

	// SequenceBits represents the number of bits in the sequence; defaults to 12
	SequenceBits uint
}

func (o Options) build() Options {
	opts := Options{
		ServerId:     o.ServerId,
		ServerBits:   o.ServerBits,
		SequenceBits: o.SequenceBits,
	}

	if o.ServerBits == 0 {
		opts.ServerBits = ServerBits
	}
	if o.SequenceBits == 0 {
		opts.SequenceBits = SequenceBits
	}

	return opts
}

// Factory is a generator of ids using Twitter's snowflake pattern
type Factory struct {
	serverId     int64
	serverBits   uint
	serverMask   int64
	sequence     int64
	sequenceBits uint
	sequenceMax  int64
	lastTime     int64
	mux          *sync.Mutex
}

func maxValue(bits uint) int64 {
	var value int64 = 1
	for i := 0; i < int(bits); i++ {
		value = value * 2
	}

	return value
}

func mask(bits uint) int64 {
	return maxValue(bits) - 1
}

// New constructs a new snowflake Factory
func New(opts Options) *Factory {
	opts = opts.build()

	serverMask := mask(opts.ServerBits)

	return &Factory{
		serverId:     opts.ServerId & serverMask,
		serverBits:   opts.ServerBits,
		serverMask:   serverMask,
		sequenceBits: opts.SequenceBits,
		sequenceMax:  maxValue(opts.SequenceBits),
		mux:          &sync.Mutex{},
	}
}

// IdN requests the next n ids
func (s *Factory) IdN(n int) []int64 {
	s.mux.Lock()
	defer s.mux.Unlock()

	var ticks time.Duration
	t := time.Now().UnixNano()/nanosInMilli - epoch
	ids := make([]int64, n)

	for i := 0; i < n; i++ {
		if t == s.lastTime {
			s.sequence = s.sequence + 1
			if s.sequence == s.sequenceMax {
				// sequence has reached it's maximum value, it's time to move to the next time slow
				s.sequence = 0
				t = t + 1
				ticks = ticks + 1
			}
		} else {
			s.sequence = 0
		}
		s.lastTime = t

		ids[i] = (t << (s.serverBits + s.sequenceBits)) | (s.serverId << s.sequenceBits) | s.sequence
	}

	if ticks > 0 {
		time.Sleep(ticks * time.Millisecond) // ensure that the clock gets advanced to avoid race conditions
	}

	return ids
}

// StringN requests the next n ids represented as a base36 string 0-9, a-z
func StringN(ids []int64) []string {
	values := make([]string, len(ids))

	for index, id := range ids {
		values[index] = strconv.FormatInt(id, 36)
	}

	return values
}
