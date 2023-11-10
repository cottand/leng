package main

import (
	"github.com/cottand/leng/internal/metric"
	"github.com/miekg/dns"
	"net"
)

type CustomDNSRecords struct {
	name   string
	answer []dns.RR
}

func NewCustomDNSRecordsFromText(recordsText []string) []CustomDNSRecords {
	customRecordsMap := make(map[string][]dns.RR)
	for _, recordText := range recordsText {
		answer, answerErr := dns.NewRR(recordText)
		if answerErr != nil {
			logger.Errorf("Cannot parse custom record: %s", answerErr)
			continue
		}
		name := answer.Header().Name
		if len(name) > 0 {
			if customRecordsMap[name] == nil {
				customRecordsMap[name] = []dns.RR{}
			}
			customRecordsMap[name] = append(customRecordsMap[name], answer)
		} else {
			logger.Errorf("Cannot parse custom record (invalid name): '%s'", recordText)
			continue
		}
	}
	return NewCustomDNSRecords(customRecordsMap)
}

func NewCustomDNSRecords(from map[string][]dns.RR) []CustomDNSRecords {
	var records []CustomDNSRecords
	var total int
	for name, rrs := range from {
		records = append(records, CustomDNSRecords{
			name:   name,
			answer: rrs,
		})
		total += len(rrs)
	}
	metric.CustomRecordCount.Set(float64(total))
	return records
}

func (records CustomDNSRecords) asHandler() func(dns.ResponseWriter, *dns.Msg) {
	return func(writer dns.ResponseWriter, req *dns.Msg) {
		m := new(dns.Msg)
		m.SetReply(req)
		m.Answer = append(m.Answer, records.answer...)

		WriteReplyMsg(writer, m)
		metric.RequestCustomCounter.Inc()
		metric.ReportDNSResponse(writer, m, false)
	}
}

// CustomRecordsResolver allows faking an in-mem DNS server just for custom records
type CustomRecordsResolver struct {
	mux *dns.ServeMux
}

func NewCustomRecordsResolver(records []CustomDNSRecords) *CustomRecordsResolver {
	mux := dns.NewServeMux()
	for _, r := range records {
		mux.HandleFunc(r.name, r.asHandler())
	}
	return &CustomRecordsResolver{mux}
}

// Resolve returns nil when there was no result found
func (r *CustomRecordsResolver) Resolve(req *dns.Msg, local net.Addr, remote net.Addr) *dns.Msg {
	writer := roResponseWriter{local: local, remote: remote}
	r.mux.ServeDNS(&writer, req)
	if writer.result.Rcode == dns.RcodeRefused {
		return nil
	} else {
		return writer.result
	}
}

// roResponseWriter implements dns.ResponseWriter,
// but does not allow calling any method with
// side effects.
// It allows wrapping a dns.ResponseWriter in order
// to recover the final written dns.Msg
type roResponseWriter struct {
	local  net.Addr
	remote net.Addr
	result *dns.Msg
}

func (w *roResponseWriter) LocalAddr() net.Addr {
	return w.local
}

func (w *roResponseWriter) RemoteAddr() net.Addr {
	return w.remote
}

func (w *roResponseWriter) WriteMsg(msg *dns.Msg) error {
	w.result = msg
	return nil
}
func (w *roResponseWriter) Write([]byte) (int, error) {
	return 0, nil
}
func (w *roResponseWriter) Close() error {
	return nil
}
func (w *roResponseWriter) TsigStatus() error {
	return nil
}
func (w *roResponseWriter) TsigTimersOnly(_ bool) {}
func (w *roResponseWriter) Hijack() {
}
