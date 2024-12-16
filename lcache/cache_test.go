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
	assert.NotNil(t, msg)

	for _, answer := range msg.Answer {
		switch answer.Header().Rrtype {
		case dns.TypeA:
			assert.Equal(t, attl, answer.Header().Ttl, "TTL should be unchanged")
		case dns.TypeAAAA:
			// AAAA now gets the TTL of A because it is smaller
			assert.Equal(t, attl, answer.Header().Ttl, "TTL should be unchanged")
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
			assert.Equal(t, attl-5, answer.Header().Ttl, "TTL should be decreased")
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
			assert.Equal(t, attl-10, answer.Header().Ttl, "TTL should be decreased")
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

	// accessing an expired key will remove it from the cache, but not straight away
	time.Sleep(1 * time.Millisecond) // allow the coro that removes the entry to run
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
	a := dns.A{
		Hdr: dns.RR_Header{
			Name:   testDomain,
			Rrtype: dns.TypeA,
			Class:  dns.ClassINET,
			Ttl:    1,
		},
		A: nullroute,
	}
	m.Answer = append(m.Answer, &a)

	if err := cache.Set(testDomain, m, true); err != nil {
		t.Error(err)
	}

	count := 10_000

	for i := 0; i < count; i++ {
		fakeClock.Advance(time.Duration(100) * time.Millisecond)
		wg := &sync.WaitGroup{}
		wg.Add(2)
		go func() {
			defer wg.Done()
			_, _, err := cache.Get(testDomain)
			if err != nil && !errors.Is(err, KeyNotFound{}) {
				t.Error(err)
			}
		}()
		go func() {
			defer wg.Done()
			newA := dns.A{
				Hdr: dns.RR_Header{
					Name:   testDomain,
					Rrtype: dns.TypeA,
					Class:  dns.ClassINET,
					Ttl:    1,
				},
				A: nullroute,
			}
			m := new(dns.Msg)
			m.SetQuestion(dns.Fqdn(testDomain), dns.TypeA)
			m.Answer = append(m.Answer, &newA)

			err := cache.Set(testDomain, m, true)
			if err != nil && !errors.Is(err, KeyNotFound{}) {
				t.Error(err)
			}
		}()
		wg.Wait()
	}
}

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
