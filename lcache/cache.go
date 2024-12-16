package lcache

import (
	"github.com/jonboulle/clockwork"
	"github.com/miekg/dns"
	"github.com/op/go-logging"
	"math"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var logger = logging.MustGetLogger("test")

// wallClock is the wall clock
var wallClock = clockwork.NewRealClock()

// entry represents a cache entry
type entry struct {
	Msg       *dns.Msg
	Blocked   bool
	expiresAt time.Time
	mu        sync.Mutex
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
	backend sync.Map // of string -> *entry
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

	existing, ok := c.backend.Load(key)
	if !ok {
		logger.Debugf("Cache: Cannot find key %s\n", key)
		return nil, false, KeyNotFound{key}
	}
	mesg := existing.(*entry)
	if mesg.Msg == nil {
		return nil, mesg.Blocked, nil
	}
	mesg.mu.Lock()
	defer mesg.mu.Unlock()
	now := wallClock.Now()

	// entry expired!
	if now.After(mesg.expiresAt) {
		c.Remove(key)
		return nil, false, KeyExpired{key}
	}
	newTtl := uint32(mesg.expiresAt.Sub(now).Truncate(time.Second).Seconds())

	for _, answer := range mesg.Msg.Answer {
		// this can happen concurrently (and it is a concurrent write of shared memory),
		// but it's ok because two concurrent modifications usually have the same result
		// when rounded to the second
		answer.Header().Ttl = newTtl
	}

	return mesg.Msg, mesg.Blocked, nil
}

func minTtlFor(msg *dns.Msg) time.Duration {
	if msg == nil {
		return 0
	}
	// find smallest ttl
	minTtl := uint32(math.MaxUint32)
	for _, answer := range msg.Answer {
		msgTtl := answer.Header().Ttl
		if minTtl > msgTtl {
			minTtl = msgTtl
		}
	}
	return time.Duration(minTtl) * time.Second
}

func (c *lengCache) Set(key string, msg *dns.Msg, blocked bool) error {
	key = strings.ToLower(key)

	if c.Full() && !c.Exists(key) {
		return CacheIsFull{}
	}
	if msg == nil {
		logger.Debugf("Setting an empty value for key %s", key)
	}

	now := wallClock.Now()
	e := entry{
		Msg:       msg,
		Blocked:   blocked,
		expiresAt: now.Add(minTtlFor(msg)),
	}
	c.backend.Store(key, &e)
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
