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

type Entry interface {
	// RRs contains dns.RR and their TTL is what
	// is used to determine entry freshness
	RRs() []dns.RR
}

type lEntry[E Entry] struct {
	underlying *E
	expiresAt  time.Time
}

// Cache interface
type Cache[E Entry] interface {
	Get(key string) (entry *E, err error)
	Set(key string, entry *E) error
	Exists(key string) bool
	Remove(key string)
	Length() int
	Full() bool
}

type lengCache[E Entry] struct {
	backend sync.Map // of string -> lEntry, which contains a *dns.Msg
	size    atomic.Int64
	full    bool
	maxSize int64
}

func NewGeneric[E Entry](maxSize int64) Cache[E] {
	return &lengCache[E]{
		backend: sync.Map{},
		size:    atomic.Int64{},
		maxSize: maxSize,
	}
}

func (c *lengCache[E]) Get(key string) (ret *E, err error) {
	key = strings.ToLower(key)

	existing, ok := c.backend.Load(key)
	if !ok {
		logger.Debugf("Cache: Cannot find key %s\n", key)
		return ret, KeyNotFound{key}
	}
	entry := existing.(lEntry[E])
	now := wallClock.Now()

	// entry expired!
	if now.After(entry.expiresAt) {
		c.Remove(key)
		return ret, KeyExpired{key}
	}
	newTtl := uint32(entry.expiresAt.Sub(now).Truncate(time.Second).Seconds())

	underlying := entry.underlying
	if underlying == nil {
		logger.Errorf("unexpected nil entry in cache")
		return nil, KeyNotFound{key}
	}
	deref := *underlying
	for _, answer := range deref.RRs() {
		// this can happen concurrently (and it is a concurrent write of shared memory),
		// but it's ok because two concurrent modifications usually have the same result
		// when rounded to the second
		answer.Header().Ttl = newTtl
	}

	return entry.underlying, nil
}

func minTtlFor[E Entry](entry *E) time.Duration {
	dereferenced := *entry
	// find smallest ttl
	minTtl := uint32(math.MaxUint32)
	for _, answer := range dereferenced.RRs() {
		msgTtl := answer.Header().Ttl
		if minTtl > msgTtl {
			minTtl = msgTtl
		}
	}
	return time.Duration(minTtl) * time.Second
}

func (c *lengCache[E]) Set(key string, entry *E) error {
	if entry == nil {
		c.Remove(key)
		return nil
	}
	key = strings.ToLower(key)

	if c.Full() && !c.Exists(key) {
		return CacheIsFull{}
	}
	now := wallClock.Now()
	e := lEntry[E]{
		underlying: entry,
		expiresAt:  now.Add(minTtlFor(entry)),
	}
	c.backend.Store(key, e)
	return nil
}

func (c *lengCache[E]) Exists(key string) bool {
	key = strings.ToLower(key)
	_, ok := c.backend.Load(key)
	return ok
}

func (c *lengCache[E]) Remove(key string) {
	_, loaded := c.backend.LoadAndDelete(key)
	if loaded {
		newSize := c.size.Add(-1)
		if newSize < c.maxSize {
			c.full = false
		}
	}
}

func (c *lengCache[E]) Length() int {
	size := c.size.Load()
	c.full = size > c.maxSize
	return int(size)
}

func (c *lengCache[E]) Full() bool {
	if c.maxSize > 0 {
		return c.full
	}
	return false
}
