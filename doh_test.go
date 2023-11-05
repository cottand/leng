package main

import (
	"github.com/miekg/dns"
	"net/http"
	"strings"
	"testing"
	"time"
)

func dnsAQuestion(question string) (msg *dns.Msg) {
	msg = new(dns.Msg)
	msg.SetQuestion(question, dns.TypeA)
	return msg
}

func TestDohHappyPath(t *testing.T) {
	handler := dns.NewServeMux()
	custom := NewCustomDNSRecordsFromText([]string{"example.com.   IN A   10.0.0.0 "})
	handler.HandleFunc("example.com", custom[0].serve(nil))

	dohTest(t, handler, func(r Resolver, bind string) {
		response, err := r.DoHLookup("http://"+bind+"/dns-query", 1, dnsAQuestion("example.com."))

		if err != nil {
			t.Fatalf("unexpected error during lookup %v", err)
		}

		if !strings.Contains(response.Answer[0].String(), "10.0.0.0") {
			t.Fatalf("failed to answer dns query for example.org - expected 10.0.0.0 but got %v", response.Answer)
		}
	})

}

func TestDoh404(t *testing.T) {
	handler := dns.NewServeMux()
	custom := NewCustomDNSRecordsFromText([]string{"example.com A 10.0.0.0"})
	handler.HandleFunc("example.com", custom[0].serve(nil))

	dohTest(t, handler, func(r Resolver, bind string) {
		resp, err := http.Get("http://" + bind + "/unknown-path")

		if resp.StatusCode != 404 {
			t.Fatalf("expected 404 but got %v", resp.StatusCode)
		}

		if err != nil {
			t.Fatalf("unexpected error during lookup %v", err)
		}
	})
}
func dohTest(t *testing.T, handler dns.Handler, doTest func(r Resolver, bind string)) {
	bind := "localhost:7698"
	config := parseDefaultConfig()
	loggingState, _ := loggerInit(config.LogConfig)
	config.DnsOverHttpServer.Bind = bind
	defer func() {
		loggingState.cleanUp()
	}()
	doh, err := NewServerHTTPS(handler, &config)
	defer doh.Shutdown()

	if err != nil {
		t.Fatalf("error when tarting server %v", err)
	}

	go func() {
		_ = doh.httpsServer.ListenAndServe()
	}()

	time.Sleep(100 * time.Millisecond)

	r := Resolver{}

	doTest(r, bind)
}
