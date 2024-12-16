package lcache

import (
	"github.com/jonboulle/clockwork"
	"github.com/miekg/dns"
	"github.com/op/go-logging"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var logger = logging.MustGetLogger("test")

// wallClock is the wall clock
var wallClock = clockwork.NewRealClock()

// Mesg represents a cache entry
type Mesg struct {
	Msg            *dns.Msg
	Blocked        bool
	LastUpdateTime time.Time
}

// Cache interface
type Cache interface {
	Get(key string) (Msg *dns.Msg, blocked bool, err error)
	Set(key string, Msg *dns.Msg, blocked bool) error
	Exists(key string) bool
	Remove(key string)
	Length() int
	Full() bool
}

type lengCache struct {
	backend sync.Map // of string -> *Mesg
	size    atomic.Int64
	full    bool
	maxSize int64
}

func New(maxSize int64) Cache {
	return &lengCache{
		backend: sync.Map{},
		size:    atomic.Int64{},
		maxSize: maxSize,
	}
}

func (c *lengCache) Get(key string) (Msg *dns.Msg, blocked bool, err error) {
	key = strings.ToLower(key)

	//Truncate time to the second, so that subsecond queries won't keep moving
	//forward the last update time without touching the TTL
	now := wallClock.Now().Truncate(time.Second)
	expired := false
	existing, ok := c.backend.Load(key)
	if !ok {
		logger.Debugf("Cache: Cannot find key %s\n", key)
		return nil, false, KeyNotFound{key}
	}
	mesg := existing.(*Mesg)
	defer func() {
		mesg.LastUpdateTime = now
	}()
	if mesg.Msg == nil {
		return nil, mesg.Blocked, nil
	}

	elapsed := uint32(now.Sub(mesg.LastUpdateTime).Seconds())
	for _, answer := range mesg.Msg.Answer {
		if elapsed > answer.Header().Ttl {
			logger.Debugf("Cache: Key expired %s", key)
			c.Remove(key)
			expired = true
		}
		answer.Header().Ttl -= elapsed
	}

	if expired {
		return nil, false, KeyExpired{key}
	}

	return mesg.Msg, mesg.Blocked, nil
}

func (c *lengCache) Set(key string, msg *dns.Msg, blocked bool) error {
	key = strings.ToLower(key)

	if c.Full() && !c.Exists(key) {
		return CacheIsFull{}
	}
	if msg == nil {
		logger.Debugf("Setting an empty value for key %s", key)
	}
	c.backend.Store(key, &Mesg{msg, blocked, wallClock.Now().Truncate(time.Second)})
	return nil
}

func (c *lengCache) Exists(key string) bool {
	key = strings.ToLower(key)
	_, ok := c.backend.Load(key)
	return ok
}

func (c *lengCache) Remove(key string) {
	_, loaded := c.backend.LoadAndDelete(key)
	if loaded {
		newSize := c.size.Add(-1)
		if newSize < c.maxSize {
			c.full = false
		}
	}
}

func (c *lengCache) Length() int {
	size := c.size.Load()
	c.full = size > c.maxSize
	return int(size)
}

func (c *lengCache) Full() bool {
	if c.maxSize > 0 {
		return c.full
	}
	return false
}
