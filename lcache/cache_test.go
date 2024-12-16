package lcache

import (
	"errors"
	"fmt"
	"github.com/jonboulle/clockwork"
	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
	"net"
	"strings"
	"sync"
	"testing"
	"time"
)

const (
	testDomain     = "www.google.com"
	testNameserver = "127.0.0.1:53"
)

func TestAdd(t *testing.T) {
	cache := New(-1)
	wallClock = clockwork.NewFakeClock()

	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(testDomain), dns.TypeA)

	if err := cache.Set(testDomain, m, false); err != nil {
		t.Error(err)
	}

	if _, _, err := cache.Get(testDomain); err != nil {
		t.Error(err)
	}

	cache.Remove(testDomain)

	if _, _, err := cache.Get(testDomain); err == nil {
		t.Error("cache entry still existed after remove")
	}
}

func TestBlockCache(t *testing.T) {
	const (
		testDomain = "www.google.com"
	)

	cache := New(-1)

	if err := cache.Set(testDomain, nil, true); err != nil {
		t.Error(err)
	}

	if exists := cache.Exists(testDomain); !exists {
		t.Error(testDomain, "didnt exist in block cache")
	}

	if exists := cache.Exists(strings.ToUpper(testDomain)); !exists {
		t.Error(strings.ToUpper(testDomain), "didnt exist in block cache")
	}

	if _, _, err := cache.Get(testDomain); err != nil {
		t.Error(err)
	}

	if exists := cache.Exists(fmt.Sprintf("%sfuzz", testDomain)); exists {
		t.Error("fuzz existed in block cache")
	}
}

func TestCacheTtl(t *testing.T) {
	fakeClock := clockwork.NewFakeClock()
	wallClock = fakeClock
	cache := New(-1)

	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(testDomain), dns.TypeA)

	var attl uint32 = 10
	var aaaattl uint32 = 20
	nullroute := net.ParseIP("0.0.0.0")
	nullroutev6 := net.ParseIP("0::0")
	a := &dns.A{
		Hdr: dns.RR_Header{
			Name:   testDomain,
			Rrtype: dns.TypeA,
			Class:  dns.ClassINET,
			Ttl:    attl,
		},
		A: nullroute}
	m.Answer = append(m.Answer, a)

	aaaa := &dns.AAAA{
		Hdr: dns.RR_Header{
			Name:   testDomain,
			Rrtype: dns.TypeAAAA,
			Class:  dns.ClassINET,
			Ttl:    aaaattl,
		},
		AAAA: nullroutev6}
	m.Answer = append(m.Answer, aaaa)

	if err := cache.Set(testDomain, m, true); err != nil {
		t.Error(err)
	}

	msg, _, err := cache.Get(testDomain)
	assert.Nil(t, err)

	for _, answer := range msg.Answer {
		switch answer.Header().Rrtype {
		case dns.TypeA:
			assert.Equal(t, attl, answer.Header().Ttl, "TTL should be unchanged")
		case dns.TypeAAAA:
			assert.Equal(t, aaaattl, answer.Header().Ttl, "TTL should be unchanged")
		default:
			t.Error("Unexpected RR type")
		}
	}

	fakeClock.Advance(5 * time.Second)
	msg, _, err = cache.Get(testDomain)
	assert.Nil(t, err)

	for _, answer := range msg.Answer {
		switch answer.Header().Rrtype {
		case dns.TypeA:
			assert.Equal(t, attl-5, answer.Header().Ttl, "TTL should be decreased")
		case dns.TypeAAAA:
			assert.Equal(t, aaaattl-5, answer.Header().Ttl, "TTL should be decreased")
		default:
			t.Error("Unexpected RR type")
		}
	}

	fakeClock.Advance(5 * time.Second)
	_, _, err = cache.Get(testDomain)
	assert.Nil(t, err)

	for _, answer := range msg.Answer {
		switch answer.Header().Rrtype {
		case dns.TypeA:
			assert.Equal(t, uint32(0), answer.Header().Ttl, "TTL should be zero")
		case dns.TypeAAAA:
			assert.Equal(t, aaaattl-10, answer.Header().Ttl, "TTL should be decreased")
		default:
			t.Error("Unexpected RR type")
		}
	}

	fakeClock.Advance(1 * time.Second)

	// accessing an expired key will return KeyExpired error
	_, _, err = cache.Get(testDomain)
	var keyExpired KeyExpired
	if !errors.As(err, &keyExpired) {
		t.Error(err)
	}

	// accessing an expired key will remove it from the cache
	_, _, err = cache.Get(testDomain)

	var keyNotFound KeyNotFound
	if !errors.As(err, &keyNotFound) {
		t.Error("cache entry still existed after expiring - ", err)
	}

}

func TestExpirationRace(t *testing.T) {
	cache := New(-1)
	fakeClock := clockwork.NewFakeClock()
	wallClock = fakeClock

	const testDomain = "www.domain.com"

	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(testDomain), dns.TypeA)

	nullroute := net.ParseIP("0.0.0.0")
	a := &dns.A{
		Hdr: dns.RR_Header{
			Name:   testDomain,
			Rrtype: dns.TypeA,
			Class:  dns.ClassINET,
			Ttl:    1,
		},
		A: nullroute}
	m.Answer = append(m.Answer, a)

	if err := cache.Set(testDomain, m, true); err != nil {
		t.Error(err)
	}

	for i := 0; i < 1000; i++ {
		fakeClock.Advance(time.Duration(100) * time.Millisecond)
		wg := &sync.WaitGroup{}
		wg.Add(2)
		go func() {
			_, _, err := cache.Get(testDomain)
			wg.Done()
			if err != nil && errors.Is(err, &KeyNotFound{}) {
				t.Error(err)
			}
		}()
		go func() {
			err := cache.Set(testDomain, m, true)
			wg.Done()
			if err != nil {
				t.Error(err)
			}
		}()
		wg.Wait()
	}
}

//func addToCache(cache Cache, time int64) {
//	cache.Add(QuestionCacheEntry{
//		Date:    time,
//		Remote:  fmt.Sprintf("%d", time),
//		Blocked: true,
//		Query:   Question{},
//	})
//}
//
//func TestQuestionCacheGetFromTimestamp(t *testing.T) {
//	memCache := makeQuestionCache(100)
//	for i := 0; i < 100; i++ {
//		addToCache(memCache, int64(i))
//	}
//
//	entries := memCache.GetOlder(50)
//	assert.Len(t, entries, 49)
//	entries = memCache.GetOlder(0)
//	assert.Len(t, entries, 99)
//	entries = memCache.GetOlder(-1)
//	assert.Len(t, entries, 100)
//	entries = memCache.GetOlder(200)
//	assert.Len(t, entries, 0)
//}

func BenchmarkSetCache(b *testing.B) {
	cache := New(-1)

	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(testDomain), dns.TypeA)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if err := cache.Set(testDomain, m, false); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGetCache(b *testing.B) {
	cache := New(-1)

	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(testDomain), dns.TypeA)

	if err := cache.Set(testDomain, m, false); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if _, _, err := cache.Get(testDomain); err != nil {
			b.Fatal(err)
		}
	}
}
