package main

import (
	"time"

	"github.com/miekg/dns"
)

// Server type
type Server struct {
	host                  string
	rTimeout              time.Duration
	wTimeout              time.Duration
	handler               *DNSHandler
	udpServer             *dns.Server
	tcpServer             *dns.Server
	udpHandler            *dns.ServeMux
	tcpHandler            *dns.ServeMux
	activeHandlerPatterns []string
}

// Run starts the server
func (s *Server) Run(
	config *Config,
	blockCache *MemoryBlockCache,
	questionCache *MemoryQuestionCache,
) {

	s.handler = NewHandler(config, blockCache, questionCache)

	tcpHandler := dns.NewServeMux()
	tcpHandler.HandleFunc(".", s.handler.DoTCP)

	udpHandler := dns.NewServeMux()
	udpHandler.HandleFunc(".", s.handler.DoUDP)

	handlerPatterns := make([]string, len(config.CustomDNSRecords))

	for _, record := range NewCustomDNSRecordsFromText(config.CustomDNSRecords) {
		dnsHandler := record.serve(s.handler)
		tcpHandler.HandleFunc(record.name, dnsHandler)
		udpHandler.HandleFunc(record.name, dnsHandler)
		handlerPatterns = append(handlerPatterns, record.name)
	}
	s.activeHandlerPatterns = handlerPatterns

	s.tcpHandler = tcpHandler
	s.udpHandler = udpHandler

	s.tcpServer = &dns.Server{
		Addr:         s.host,
		Net:          "tcp",
		Handler:      tcpHandler,
		ReadTimeout:  s.rTimeout,
		WriteTimeout: s.wTimeout,
	}

	s.udpServer = &dns.Server{
		Addr:         s.host,
		Net:          "udp",
		Handler:      udpHandler,
		UDPSize:      65535,
		ReadTimeout:  s.rTimeout,
		WriteTimeout: s.wTimeout,
	}

	go s.start(s.udpServer)
	go s.start(s.tcpServer)
}

func (s *Server) start(ds *dns.Server) {
	logger.Criticalf("start %s listener on %s\n", ds.Net, s.host)

	if err := ds.ListenAndServe(); err != nil {
		logger.Criticalf("start %s listener on %s failed: %s\n", ds.Net, s.host, err.Error())
	}
}

// Stop stops the server
func (s *Server) Stop() {
	if s.handler != nil {
		s.handler.muActive.Lock()
		s.handler.active = false
		close(s.handler.requestChannel)
		s.handler.muActive.Unlock()
	}
	if s.udpServer != nil {
		err := s.udpServer.Shutdown()
		if err != nil {
			logger.Critical(err)
		}
	}
	if s.tcpServer != nil {
		err := s.tcpServer.Shutdown()
		if err != nil {
			logger.Critical(err)
		}
	}
}

// ReloadConfig only supports reloading the customDnsRecords section of the config for now
func (s *Server) ReloadConfig(config *Config) {
	oldRecords := s.activeHandlerPatterns
	newRecords := NewCustomDNSRecordsFromText(config.CustomDNSRecords)
	newRecordsPatterns := make([]string, len(newRecords))
	for _, r := range newRecords {
		newRecordsPatterns = append(newRecordsPatterns, r.name)
	}
	deletedRecords := difference(oldRecords, newRecordsPatterns)

	for _, deleted := range deletedRecords {
		s.tcpHandler.HandleRemove(deleted)
		s.udpHandler.HandleRemove(deleted)
	}

	for _, record := range newRecords {
		dnsHandler := record.serve(s.handler)
		s.tcpHandler.HandleFunc(record.name, dnsHandler)
		s.udpHandler.HandleFunc(record.name, dnsHandler)
	}
	s.activeHandlerPatterns = newRecordsPatterns
}
