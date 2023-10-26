package main

import (
	"fmt"
	"github.com/pelletier/go-toml/v2"
	"strings"
	"testing"
	"time"

	"github.com/miekg/dns"
)

func BenchmarkResolver(b *testing.B) {
	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(testDomain), dns.TypeA)

	c := new(dns.Client)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _, err := c.Exchange(m, testNameserver)
		if err != nil {
			logger.Error(err)
		}
	}
}

func integrationTest(changeConfig func(c *Config), test func(client *dns.Client, target string)) {
	testDnsHost := "127.0.0.1:5300"
	var config Config
	_ = toml.Unmarshal([]byte(defaultConfig), &config)

	changeConfig(&config)

	quitActivation := make(chan bool)
	actChannel := make(chan *ActivationHandler)

	go startActivation(actChannel, quitActivation, config.ReactivationDelay)
	grimdActivation = <-actChannel
	close(actChannel)

	server := &Server{
		host:     testDnsHost,
		rTimeout: 5 * time.Second,
		wTimeout: 5 * time.Second,
	}
	c := new(dns.Client)

	// BlockCache contains all blocked domains
	blockCache := &MemoryBlockCache{Backend: make(map[string]bool)}
	// QuestionCache contains all queries to the dns server
	questionCache := makeQuestionCache(config.QuestionCacheCap)

	server.Run(&config, blockCache, questionCache)

	time.Sleep(200 * time.Millisecond)
	defer server.Stop()

	test(c, testDnsHost)
}

func TestMultipleARecords(t *testing.T) {
	integrationTest(
		func(c *Config) {
			c.CustomDNSRecords = []string{
				"example.com.          IN  A       10.10.0.1 ",
				"example.com.          IN  A       10.10.0.2 ",
			}
		},
		func(client *dns.Client, target string) {
			c := new(dns.Client)

			// BlockCache contains all blocked domains
			m := new(dns.Msg)
			m.SetQuestion(dns.Fqdn("example.com"), dns.TypeA)

			reply, _, err := c.Exchange(m, target)
			if err != nil {
				t.Error(err)
				t.FailNow()
			}

			if l := len(reply.Answer); l != 2 {
				t.Fatalf("Expected 2 returned records but had %v: %v", l, reply.Answer)
			}
		},
	)
}

func Test2DifferentARecords(t *testing.T) {
	integrationTest(
		func(c *Config) {
			c.CustomDNSRecords = []string{
				"example.com          IN  A       10.10.0.1 ",
				"boo.org              IN  A       10.10.0.2 ",
			}
		},
		func(client *dns.Client, target string) {
			c := new(dns.Client)

			m := new(dns.Msg)

			m.SetQuestion(dns.Fqdn("example.com"), dns.TypeA)
			reply, _, err := c.Exchange(m, target)
			if err != nil {
				t.Error(err)
				t.FailNow()
			}
			if l := len(reply.Answer); l != 1 {
				t.Fatalf("Expected 1 returned records but had %v: %v", l, reply.Answer)
			}

			if !strings.Contains(reply.Answer[0].String(), "10.10.0.1") {
				t.Fatalf("Expected the right A address to be returned, but got %v", reply.Answer[0])
			}
		},
	)
}
func Test2in3DifferentARecords(t *testing.T) {
	integrationTest(
		func(c *Config) {
			c.CustomDNSRecords = []string{
				"example.com          IN  A       10.10.0.1 ",
				"boo.org              IN  A       10.10.0.2 ",
				"boo.org              IN  A       10.10.0.3 ",
			}
		},
		func(client *dns.Client, target string) {
			c := new(dns.Client)

			m := new(dns.Msg)

			m.SetQuestion(dns.Fqdn("example.com"), dns.TypeA)
			reply, _, err := c.Exchange(m, target)
			if err != nil {
				t.Error(err)
				t.FailNow()
			}
			if l := len(reply.Answer); l != 1 {
				t.Fatalf("Expected 1 returned records but had %v: %v", l, reply.Answer)
			}

			if !strings.Contains(reply.Answer[0].String(), "10.10.0.1") {
				t.Fatalf("Expected the right A address to be returned, but got %v", reply.Answer[0])
			}
		},
	)
}

func TestConfigReloadForCustomRecords(t *testing.T) {
	testDnsHost := "127.0.0.1:5300"
	var config Config
	_ = toml.Unmarshal([]byte(defaultConfig), &config)

	config.CustomDNSRecords = []string{
		// custom TLD so that we do not fall back to querying upstream DNS if missing
		"old.com_custom       IN  A       10.10.0.1 ",
		"boo.org              IN  A       10.10.0.2 ",
		"boo.org              IN  A       10.10.0.3 ",
	}

	quitActivation := make(chan bool)
	actChannel := make(chan *ActivationHandler)

	go startActivation(actChannel, quitActivation, config.ReactivationDelay)
	grimdActivation = <-actChannel
	close(actChannel)

	server := &Server{
		host:     testDnsHost,
		rTimeout: 5 * time.Second,
		wTimeout: 5 * time.Second,
	}
	c := new(dns.Client)

	// BlockCache contains all blocked domains
	blockCache := &MemoryBlockCache{Backend: make(map[string]bool)}
	// QuestionCache contains all queries to the dns server
	questionCache := makeQuestionCache(config.QuestionCacheCap)

	server.Run(&config, blockCache, questionCache)

	time.Sleep(200 * time.Millisecond)
	defer server.Stop()

	m1 := new(dns.Msg)
	m1.SetQuestion(dns.Fqdn("old.com_custom"), dns.TypeA)
	reply, _, err := c.Exchange(m1, testDnsHost)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if l := len(reply.Answer); l != 1 {
		t.Fatalf("Expected 1 returned records but had %v: %v", l, reply.Answer)
	}

	if !strings.Contains(reply.Answer[0].String(), "10.10.0.1") {
		t.Fatalf("Expected the right A address to be returned, but got %v", reply.Answer[0])
	}

	newConfig := config
	newConfig.CustomDNSRecords = []string{
		// old.com is gone, boo.org has changed
		"new.com              IN  A       10.10.0.1 ",
		"boo.org              IN  A       10.10.2.2 ",
	}

	server.ReloadConfig(&newConfig)
	time.Sleep(200 * time.Millisecond)

	m1 = new(dns.Msg)
	m1.SetQuestion(dns.Fqdn("old.com_custom"), dns.TypeA)
	reply, _, err = c.Exchange(m1, testDnsHost)
	if err != nil {
		fmt.Printf("Err was %v - expected this", err)
		t.FailNow()
	}
	if len(reply.Answer) != 0 {
		t.Fatalf("expected old.com_custom DNS to fail, but got %v", reply)
	}

	m1 = new(dns.Msg)
	m1.SetQuestion(dns.Fqdn("boo.org"), dns.TypeA)
	reply, _, err = c.Exchange(m1, testDnsHost)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if l := len(reply.Answer); l != 1 {
		t.Fatalf("Expected 1 returned records but had %v: %v", l, reply.Answer)
	}

	if !strings.Contains(reply.Answer[0].String(), "10.10.2.2") {
		t.Fatalf("Expected the new A address to be returned, but got %v", reply.Answer[0])
	}

	m1 = new(dns.Msg)
	m1.SetQuestion(dns.Fqdn("new.com"), dns.TypeA)
	reply, _, err = c.Exchange(m1, testDnsHost)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if l := len(reply.Answer); l != 1 {
		t.Fatalf("Expected 1 returned records but had %v: %v", l, reply.Answer)
	}

	if !strings.Contains(reply.Answer[0].String(), "10.10.0.1") {
		t.Fatalf("Expected the new A address to be returned, but got %v", reply.Answer[0])
	}
}
