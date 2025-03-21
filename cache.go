package main

import (
	"crypto/md5"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/miekg/dns"
	"github.com/ryanuber/go-glob"
)

const globChars = "*?"

// KeyNotFound type
type KeyNotFound struct {
	key string
}

// Error formats an error for the KeyNotFound type
func (e KeyNotFound) Error() string {
	return e.key + " " + "not found"
}

// KeyExpired type
type KeyExpired struct {
	Key string
}

// Error formats an error for the KeyExpired type
func (e KeyExpired) Error() string {
	return e.Key + " " + "expired"
}

// CacheIsFull type
type CacheIsFull struct {
}

// Error formats an error for the CacheIsFull type
func (e CacheIsFull) Error() string {
	return "Cache is Full"
}

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

func NewCache(maxSize int64) Cache {
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
	now := WallClock.Now().Truncate(time.Second)
	expired := false
	existing, ok := c.backend.Load(key)
	mesg := existing.(*Mesg)
	if ok && mesg.Msg == nil {
		ok = false
		logger.Warningf("Cache: key %s returned nil entry", key)
		c.Remove(key)
	}

	if ok {
		elapsed := uint32(now.Sub(mesg.LastUpdateTime).Seconds())
		for _, answer := range mesg.Msg.Answer {
			if elapsed > answer.Header().Ttl {
				logger.Debugf("Cache: Key expired %s", key)
				c.Remove(key)
				expired = true
			}
			answer.Header().Ttl -= elapsed
		}
	}

	if !ok {
		logger.Debugf("Cache: Cannot find key %s\n", key)
		return nil, false, KeyNotFound{key}
	}

	if expired {
		return nil, false, KeyExpired{key}
	}

	mesg.LastUpdateTime = now

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
	c.backend.Store(key, &Mesg{msg, blocked, WallClock.Now().Truncate(time.Second)})
	return nil
}

func (c *lengCache) Exists(key string) bool {
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

// MemoryCache type
type MemoryCache struct {
	Backend  map[string]*Mesg
	Maxcount int
	mu       sync.RWMutex
}

// MemoryBlockCache type
type MemoryBlockCache struct {
	Backend map[string]bool
	Special map[string]*regexp.Regexp
	mu      sync.RWMutex
}

// Get returns the entry for a key or an error
func (c *MemoryCache) Get(key string) (*dns.Msg, bool, error) {
	key = strings.ToLower(key)

	//Truncate time to the second, so that subsecond queries won't keep moving
	//forward the last update time without touching the TTL
	now := WallClock.Now().Truncate(time.Second)

	expired := false
	c.mu.Lock()
	mesg, ok := c.Backend[key]
	if ok && mesg.Msg == nil {
		ok = false
		logger.Warningf("Cache: key %s returned nil entry", key)
		c.removeNoLock(key)
	}
	if ok {
		elapsed := uint32(now.Sub(mesg.LastUpdateTime).Seconds())
		for _, answer := range mesg.Msg.Answer {
			if elapsed > answer.Header().Ttl {
				logger.Debugf("Cache: Key expired %s", key)
				c.removeNoLock(key)
				expired = true
			}
			answer.Header().Ttl -= elapsed
		}
	}
	c.mu.Unlock()

	if !ok {
		logger.Debugf("Cache: Cannot find key %s\n", key)
		return nil, false, KeyNotFound{key}
	}

	if expired {
		return nil, false, KeyExpired{key}
	}

	mesg.LastUpdateTime = now

	return mesg.Msg, mesg.Blocked, nil
}

// Set sets a keys value to a Mesg
func (c *MemoryCache) Set(key string, msg *dns.Msg, blocked bool) error {
	key = strings.ToLower(key)

	if c.Full() && !c.Exists(key) {
		return CacheIsFull{}
	}
	if msg == nil {
		logger.Debugf("Setting an empty value for key %s", key)
	}
	c.mu.Lock()
	c.Backend[key] = &Mesg{msg, blocked, WallClock.Now().Truncate(time.Second)}
	c.mu.Unlock()

	return nil
}

// Remove removes an entry from the cache
func (c *MemoryCache) removeNoLock(key string) {
	key = strings.ToLower(key)
	delete(c.Backend, key)
}

// Remove removes an entry from the cache
func (c *MemoryCache) Remove(key string) {
	c.mu.Lock()
	c.removeNoLock(key)
	c.mu.Unlock()
}

// Exists returns whether a key exists in the cache
func (c *MemoryCache) Exists(key string) bool {
	key = strings.ToLower(key)

	c.mu.RLock()
	_, ok := c.Backend[key]
	c.mu.RUnlock()
	return ok
}

// Length returns the caches length
func (c *MemoryCache) Length() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.Backend)
}

// Full returns whether the cache is full
func (c *MemoryCache) Full() bool {
	if c.Maxcount == 0 {
		return false
	}
	return c.Length() >= c.Maxcount
}

// KeyGen generates a key for the hash from a question
func KeyGen(q Question) string {
	h := md5.New()
	h.Write([]byte(q.String()))
	x := h.Sum(nil)
	key := fmt.Sprintf("%x", x)
	logger.Debugf("KeyGen: %s %s", q.String(), key)
	return key
}

// Get returns the entry for a key or an error
func (c *MemoryBlockCache) Get(key string) (bool, error) {
	var ok, val bool

	c.mu.RLock()
	if strings.HasPrefix(key, "~") {
		_, ok = c.Special[key]
		val = true
	} else if strings.ContainsAny(key, globChars) {
		key = strings.ToLower(key)
		_, ok = c.Special[key]
		val = true
	} else {
		key = strings.ToLower(key)
		val, ok = c.Backend[key]
	}
	c.mu.RUnlock()

	if !ok {
		return false, KeyNotFound{key}
	}

	return val, nil
}

// Remove removes an entry from the cache
func (c *MemoryBlockCache) Remove(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if strings.HasPrefix(key, "~") {
		delete(c.Special, key)
	} else if strings.ContainsAny(key, globChars) {
		delete(c.Special, strings.ToLower(key))
	} else {
		delete(c.Backend, strings.ToLower(key))
	}
}

// Set sets a value in the BlockCache
func (c *MemoryBlockCache) Set(key string, value bool) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if strings.HasPrefix(key, "~") {
		// get rid of the ~ prefix
		runes := []rune(key)
		ex := string(runes[1:])
		re, err := regexp.Compile(ex)
		if err != nil {
			logger.Errorf("Invalid regexp entry: `%s` %v", key, err)
		} else {
			c.Special[key] = re
		}
	} else if strings.ContainsAny(key, globChars) {
		c.Special[strings.ToLower(key)] = nil
	} else {
		c.Backend[strings.ToLower(key)] = value
	}
	return nil
}

// Exists returns whether a key exists in the cache
func (c *MemoryBlockCache) Exists(key string) bool {
	key = strings.ToLower(key)

	c.mu.RLock()
	defer c.mu.RUnlock()
	_, ok := c.Backend[key]
	if !ok {
		for data, regex := range c.Special {
			if regex != nil {
				if regex.MatchString(key) {
					ok = true
				}
			} else {
				if glob.Glob(data, key) {
					ok = true
				}
			}
		}
	}
	return ok
}

// Length returns the caches length
func (c *MemoryBlockCache) Length() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.Backend)
}
