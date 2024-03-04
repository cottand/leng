package main

import (
	"github.com/cottand/leng/internal/metric"
	"net"
	"slices"
	"sync"
	"time"

	"github.com/miekg/dns"
)

const (
	notIPQuery = 0
	_IP4Query  = 4
	_IP6Query  = 6
)

// Question type
type Question struct {
	Qname  string `json:"name"`
	Qtype  string `json:"type"`
	Qclass string `json:"class"`
}

// QuestionCacheEntry represents a full query from a client with metadata
type QuestionCacheEntry struct {
	Date    int64    `json:"date"`
	Remote  string   `json:"client"`
	Blocked bool     `json:"blocked"`
	Query   Question `json:"query"`
}

// String formats a question
func (q *Question) String() string {
	return q.Qname + " " + q.Qclass + " " + q.Qtype
}

// EventLoop type
type EventLoop struct {
	requestChannel chan DNSOperationData
	resolver       *Resolver
	cache          Cache
	// negCache caches failures
	negCache      Cache
	active        bool
	muActive      sync.RWMutex
	config        *Config
	blockCache    *MemoryBlockCache
	questionCache *MemoryQuestionCache
	customDns     *CustomRecordsResolver
}

// DNSOperationData type
type DNSOperationData struct {
	Net string
	w   dns.ResponseWriter
	req *dns.Msg
}

// NewEventLoop returns a new eventLoop
func NewEventLoop(config *Config, blockCache *MemoryBlockCache, questionCache *MemoryQuestionCache) *EventLoop {
	var (
		clientConfig *dns.ClientConfig
		resolver     *Resolver
		cache        Cache
		negCache     Cache
	)

	resolver = &Resolver{clientConfig}

	cache = &MemoryCache{
		Backend:  make(map[string]*Mesg, config.Upstream.Maxcount),
		Maxcount: config.Upstream.Maxcount,
	}
	negCache = &MemoryCache{
		Backend:  make(map[string]*Mesg),
		Maxcount: config.Upstream.Maxcount,
	}

	handler := &EventLoop{
		requestChannel: make(chan DNSOperationData),
		resolver:       resolver,
		cache:          cache,
		negCache:       negCache,
		blockCache:     blockCache,
		questionCache:  questionCache,
		active:         true,
		config:         config,
		customDns:      NewCustomRecordsResolver(NewCustomDNSRecordsFromText(config.CustomDNSRecords)),
	}

	go handler.do()

	return handler
}

func (h *EventLoop) do() {
	for {
		data, ok := <-h.requestChannel
		if !ok {
			break
		}
		h.doRequest(data.Net, data.w, data.req)
	}
}

// responseFor has side-effects, like writing to h's caches, so avoid calling it concurrently
func (h *EventLoop) responseFor(Net string, req *dns.Msg, _local net.Addr, _remote net.Addr) (_ *dns.Msg, success bool) {

	var remote net.IP
	if Net == "tcp" || Net == "http" {
		remote = _remote.(*net.TCPAddr).IP
	} else {
		remote = _remote.(*net.UDPAddr).IP
	}

	// first of all, check custom DNS. No need to cache it because it is already in-mem and precedes the blocking
	if custom := h.customDns.Resolve(req, _local, _remote); custom != nil {
		return custom, true
	}

	q := req.Question[0]
	Q := Question{UnFqdn(q.Name), dns.TypeToString[q.Qtype], dns.ClassToString[q.Qclass]}
	logger.Infof("%s lookup　%s\n", remote, Q.String())

	IPQuery := h.isIPQuery(q)

	// Only query cache when qtype == 'A'|'AAAA' , qclass == 'IN'
	key := KeyGen(Q)
	if IPQuery > 0 {
		mesg, blocked, err := h.cache.Get(key)
		if err != nil {
			if mesg, blocked, err = h.negCache.Get(key); err != nil {
				logger.Debugf("%s didn't hit cache\n", Q.String())
			} else {
				logger.Debugf("%s hit negative cache\n", Q.String())
				return nil, false
			}
		} else {
			if blocked && !lengActive {
				logger.Debugf("%s hit cache and was blocked: forwarding request\n", Q.String())
			} else {
				logger.Debugf("%s hit cache\n", Q.String())

				// we need this copy against concurrent modification of ID
				msg := *mesg
				msg.Id = req.Id

				defer metric.ReportDNSRespond(remote, &msg, blocked)
				return &msg, true
			}
		}
	}
	// Check blocklist
	var blacklisted = false

	if IPQuery > 0 {
		blacklisted = h.blockCache.Exists(Q.Qname)

		if lengActive && blacklisted {
			m := new(dns.Msg)
			m.SetReply(req)

			if h.config.Blocking.NXDomain {
				m.SetRcode(req, dns.RcodeNameError)
			} else {
				nullroute := net.ParseIP(h.config.Blocking.Nullroute)
				nullroutev6 := net.ParseIP(h.config.Blocking.Nullroutev6)

				switch IPQuery {
				case _IP4Query:
					rrHeader := dns.RR_Header{
						Name:   q.Name,
						Rrtype: dns.TypeA,
						Class:  dns.ClassINET,
						Ttl:    h.config.TTL,
					}
					a := &dns.A{Hdr: rrHeader, A: nullroute}
					m.Answer = append(m.Answer, a)
				case _IP6Query:
					rrHeader := dns.RR_Header{
						Name:   q.Name,
						Rrtype: dns.TypeAAAA,
						Class:  dns.ClassINET,
						Ttl:    h.config.TTL,
					}
					a := &dns.AAAA{Hdr: rrHeader, AAAA: nullroutev6}
					m.Answer = append(m.Answer, a)
				}
			}

			defer metric.ReportDNSRespond(remote, m, true)

			logger.Noticef("%s found in blocklist\n", Q.Qname)

			// log query
			NewEntry := QuestionCacheEntry{Date: time.Now().Unix(), Remote: remote.String(), Query: Q, Blocked: true}
			go h.questionCache.Add(NewEntry)

			// cache the block; we don't know the true TTL for blocked entries: we just enforce our config
			err := h.cache.Set(key, m, true)
			if err != nil {
				logger.Errorf("Set %s block cache failed: %s\n", Q.String(), err.Error())
			}

			return m, true
		}
		logger.Debugf("%s not found in blocklist\n", Q.Qname)
	}

	// log query
	NewEntry := QuestionCacheEntry{Date: time.Now().Unix(), Remote: remote.String(), Query: Q, Blocked: false}
	go h.questionCache.Add(NewEntry)

	mesg, err := h.resolver.Lookup(Net, req, h.config.Timeout, h.config.Interval, h.config.Upstream.Nameservers, h.config.Upstream.DoH)

	if err != nil {
		logger.Errorf("resolve query error %s\n", err)

		// cache the failure, too!
		if err = h.negCache.Set(key, nil, false); err != nil {
			logger.Errorf("set %s negative cache failed: %v\n", Q.String(), err)
		}
		return nil, false
	}

	if mesg.Truncated && Net == "udp" {
		mesg, err = h.resolver.Lookup("tcp", req, h.config.Timeout, h.config.Interval, h.config.Upstream.Nameservers, h.config.Upstream.DoH)
		if err != nil {
			logger.Errorf("resolve tcp query error %s\n", err)

			// cache the failure, too!
			if err = h.negCache.Set(key, nil, false); err != nil {
				logger.Errorf("set %s negative cache failed: %v\n", Q.String(), err)
			}
			return nil, false
		}
	}

	//find the smallest ttl
	ttl := h.config.Upstream.Expire
	var candidateTTL uint32

	for index, answer := range mesg.Answer {
		logger.Debugf("Answer %d - %s\n", index, answer.String())

		candidateTTL = answer.Header().Ttl

		if candidateTTL > 0 && candidateTTL < ttl {
			ttl = candidateTTL
		}
	}

	defer metric.ReportDNSRespond(remote, mesg, false)

	if IPQuery > 0 && len(mesg.Answer) > 0 {
		if !lengActive && blacklisted {
			logger.Debugf("%s is blacklisted and leng not active: not caching\n", Q.String())
		} else {
			err = h.cache.Set(key, mesg, false)
			if err != nil {
				logger.Errorf("set %s cache failed: %s\n", Q.String(), err.Error())
			}
			logger.Debugf("insert %s into cache with ttl %d\n", Q.String(), ttl)
		}
	}
	return mesg, true
}

func (h *EventLoop) doRequest(Net string, w dns.ResponseWriter, req *dns.Msg) {
	defer func(w dns.ResponseWriter) {
		err := w.Close()
		if err != nil {
		}
	}(w)

	resp, ok := h.responseFor(Net, req, w.LocalAddr(), w.RemoteAddr())

	if !ok {
		m := new(dns.Msg)
		m.SetRcode(req, dns.RcodeServerFailure)
		WriteReplyMsg(w, m)
		metric.ReportDNSResponse(w, m, false)
		return
	}

	depthSoFar := uint32(0)
	for h.config.FollowCnameDepth > depthSoFar {
		cnames, ok := canFollow(req, resp)
		depthSoFar++
		if !ok {
			break
		}
		for _, cname := range cnames {
			r := dns.Msg{}
			r.SetQuestion(cname.Target, req.Question[0].Qtype)
			followed, ok := h.responseFor(Net, &r, w.LocalAddr(), w.RemoteAddr())
			for _, fAnswer := range followed.Answer {
				containsNewAnswer := func(rr dns.RR) bool {
					return rr.String() == fAnswer.String()
				}
				if ok && !slices.ContainsFunc(resp.Answer, containsNewAnswer) {
					resp.Answer = append(resp.Answer, fAnswer)
				}
			}
		}
	}

	WriteReplyMsg(w, resp)

}

// determines if resp contains no A records but some CNAME record
func canFollow(req *dns.Msg, resp *dns.Msg) (cnames []*dns.CNAME, ok bool) {
	// RFC-1034: only follow non-CNAME queries
	if req.Question[0].Qtype == dns.TypeCNAME {
		return []*dns.CNAME{}, false
	}

	isA := func(rr dns.RR) bool {
		return rr.Header().Rrtype == dns.TypeA || rr.Header().Rrtype == dns.TypeAAAA
	}

	isCname := func(rr dns.RR) bool {
		return rr.Header().Rrtype == dns.TypeCNAME
	}

	ok = !slices.ContainsFunc(resp.Answer, isA) && slices.ContainsFunc(resp.Answer, isCname)
	for _, rr := range resp.Answer {
		if asCname, ok := rr.(*dns.CNAME); isCname(rr) && ok {
			cnames = append(cnames, asCname)
		}
	}

	return cnames, ok && len(cnames) != 0
}

// msg:
// Q: A   fst.com
// A: CN  snd.com, thrd.com
//

// DoTCP begins a tcp query
func (h *EventLoop) DoTCP(w dns.ResponseWriter, req *dns.Msg) {
	h.muActive.RLock()
	defer h.muActive.RUnlock()
	if h.active {
		h.requestChannel <- DNSOperationData{"tcp", w, req}
	}
}

// DoUDP begins a udp query
func (h *EventLoop) DoUDP(w dns.ResponseWriter, req *dns.Msg) {
	h.muActive.RLock()
	defer h.muActive.RUnlock()
	if h.active {
		h.requestChannel <- DNSOperationData{"udp", w, req}
	}
}

func (h *EventLoop) DoHTTP(w dns.ResponseWriter, req *dns.Msg) {
	h.muActive.RLock()
	defer h.muActive.RUnlock()
	if h.active {
		h.requestChannel <- DNSOperationData{"http", w, req}
	}
}

// WriteReplyMsg writes the dns reply
func WriteReplyMsg(w dns.ResponseWriter, message *dns.Msg) {
	defer func() {
		if r := recover(); r != nil {
			logger.Noticef("Recovered in WriteReplyMsg: %s\n", r)
		}
	}()

	err := w.WriteMsg(message)
	if err != nil {
		logger.Error(err)
	}
}

func (h *EventLoop) isIPQuery(q dns.Question) int {
	if q.Qclass != dns.ClassINET {
		return notIPQuery
	}

	switch q.Qtype {
	case dns.TypeA:
		return _IP4Query
	case dns.TypeAAAA:
		return _IP6Query
	default:
		return notIPQuery
	}
}

// UnFqdn function
func UnFqdn(s string) string {
	if dns.IsFqdn(s) {
		return s[:len(s)-1]
	}
	return s
}
