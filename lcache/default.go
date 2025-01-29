package lcache

import "github.com/miekg/dns"

var _ Entry = DefaultEntry{}

// DefaultEntry is the default implementation of Entry
type DefaultEntry struct {
	dns.Msg
}

func (dnsEntry DefaultEntry) RRs() []dns.RR {
	return dnsEntry.Answer
}

func NewDefault(maxSize int) Cache[DefaultEntry] {
	return NewGeneric[DefaultEntry](maxSize)
}
