package reqLogic

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"github.com/i5heu/PathfinderBeacon/pkg/utils"
	"github.com/miekg/dns"
	"go.uber.org/zap"
)

func generateCookie() string {
	cookie := make([]byte, 64)
	_, err := rand.Read(cookie)
	if err != nil {
		log.Fatalf("Failed to generate cookie: %s\n", err)
	}

	// return hex.EncodeToString(cookie)
	return "0a57a6d8fa081b89"
}

func checkForEdnsCookie(r *dns.Msg) string {
	if opt := r.IsEdns0(); opt != nil {
		for _, o := range opt.Option {
			if cookie, ok := o.(*dns.EDNS0_COOKIE); ok {
				return cookie.Cookie
			}
		}
	}
	return ""
}

func setEdns0AndCookieAndCreateMsg(w dns.ResponseWriter, r *dns.Msg) *dns.Msg {
	msg := new(dns.Msg)
	msg.SetReply(r)
	msg.Authoritative = true

	// optRecord := &dns.OPT{
	// 	Hdr: dns.RR_Header{
	// 		Name:   ".",
	// 		Rrtype: dns.TypeOPT,
	// 	},
	// 	Option: []dns.EDNS0{
	// 		&dns.EDNS0_COOKIE{
	// 			Code:   dns.EDNS0COOKIE,
	// 			Cookie: checkForEdnsCookie(r) + generateCookie(),
	// 		},
	// 	},
	// }
	// optRecord.SetUDPSize(4096)
	// optRecord.SetDo()
	// msg.Extra = append(msg.Extra, optRecord)
	return msg
}

func moveToTCP(msg *dns.Msg, w dns.ResponseWriter, r *dns.Msg) {
	msg.SetRcode(r, dns.RcodeSuccess)
	msg.Truncated = true
	w.WriteMsg(msg)
}

func (d *ReqLogic) DNSReq(w dns.ResponseWriter, r *dns.Msg) {
	start := time.Now()
	ctx := context.Background()

	var err error
	if IsUDPRequest(w.RemoteAddr()) {
		_, _, _, ok, err := d.globalRateLimitStoreUDP.Take(ctx, getIPFromRemoteAddr(w.RemoteAddr().String()))
		if err != nil {
			log.Printf("Failed to get global rate limit: %s\n", err)
		}
		if !ok {
			log.Printf("Global rate limit exceeded for %s\n", w.RemoteAddr().String())
			return
		}
	}

	var msg *dns.Msg

	if IsUDPRequest(w.RemoteAddr()) {
		msg = setEdns0AndCookieAndCreateMsg(w, r)
	} else {
		msg = new(dns.Msg)
		msg.SetReply(r)
		msg.Authoritative = true
	}

	defer func(start time.Time) {
		ctxClose := context.Background()
		tokens, remaining, errR := d.rateLimitStoreUDP.Get(ctxClose, getIPFromRemoteAddr(w.RemoteAddr().String()))
		if errR != nil {
			fmt.Printf("Failed to get rate limit: %s\n", errR)
		}
		d.logger.Info("Request",
			zap.String("remote_addr", w.RemoteAddr().String()),
			zap.Bool("UDP", IsUDPRequest(w.RemoteAddr())),
			zap.Duration("duration", time.Since(start)),
			zap.Uint64("rate_limit_tokens", tokens),
			zap.Uint64("rate_limit_remaining", remaining),
			zap.Any("question", r.Question),
			zap.Error(err))
	}(start)

	for _, q := range r.Question {
		err = rateLimit(ctx, w, d)
		if err != nil {
			return
		}
		if utils.ToLowerCase(q.Name) != "pathfinderbeacon.net." && !strings.HasSuffix(utils.ToLowerCase(q.Name), ".pathfinderbeacon.net.") && !strings.HasSuffix(utils.ToLowerCase(q.Name), ".heidenstedt.org.") {
			msg.Rcode = dns.RcodeNameError
			break
		}

		if !strings.HasSuffix("a."+utils.ToLowerCase(q.Name), ".q.") {
			err = rateLimit(ctx, w, d)
			if err != nil {
				return
			}
		}

		switch q.Qtype {
		case dns.TypeSOA:
			handleSOARequest(msg, q)
		case dns.TypeTXT:
			// If the request is a UDP request, move it to TCP if it is a TXT request
			if IsUDPRequest(w.RemoteAddr()) {
				moveToTCP(msg, w, r)
				return
			}
			d.handleTXTRequest(msg, q)
		case dns.TypeNS:
			handleNSRequest(msg, q)
		case dns.TypeA:
			handleARequest(msg, q)
		case dns.TypeAAAA:
			handleAAAARequest(msg, q)
		case dns.TypeCAA:
		default:
			handleNotImplemented(msg)
		}
	}

	err = w.WriteMsg(msg)
	if err != nil {
		d.logger.Info("Failed to write message", zap.Error(err),
			zap.String("remote_addr", w.RemoteAddr().String()),
			zap.Any("question", r.Question),
			zap.String("answer", msg.String()))
	}
}

// Additional DNS request handlers (handleSOARequest, handleARequest, handleAAAARequest, etc.)
func handleSOARequest(msg *dns.Msg, q dns.Question) {
	soa := &dns.SOA{
		Hdr: dns.RR_Header{
			Name:   utils.ToLowerCase(q.Name),
			Rrtype: dns.TypeSOA,
			Class:  dns.ClassINET,
			Ttl:    300,
		},
		Ns:      "pathfinderbeacon-ns1.heidenstedt.org.",
		Mbox:    "hostmaster-pathfinderbeacon-net.heidenstedt.org.",
		Serial:  uint32(time.Now().Unix()),
		Refresh: 7200,
		Retry:   3600,
		Expire:  1209600,
		Minttl:  300,
	}
	msg.Answer = append(msg.Answer, soa)
}

func handleNSRequest(msg *dns.Msg, q dns.Question) {
	ns := &dns.NS{
		Hdr: dns.RR_Header{
			Name:   utils.ToLowerCase(q.Name),
			Rrtype: dns.TypeNS,
			Class:  dns.ClassINET,
			Ttl:    300,
		},
		Ns: "pathfinderbeacon-ns1.heidenstedt.org.",
	}
	msg.Answer = append(msg.Answer, ns)
}

func (d *ReqLogic) handleTXTRequest(msg *dns.Msg, q dns.Question) {

	var requestType string
	switch {
	case strings.HasSuffix(q.Name, "room.pathfinderbeacon.net."):
		requestType = "room"
	case strings.HasSuffix(q.Name, "node.pathfinderbeacon.net."):
		requestType = "node"
	case strings.HasSuffix(q.Name, "auth.pathfinderbeacon.net."):
		handleTxtAuthRequest(msg, q)
		return

	default:
		return
	}

	name, ok := utils.GetDNSParameterAndCheckIfSha224(q.Name)
	if !ok {
		msg.Rcode = dns.RcodeNameError
		return
	}

	values, err := d.GetValues(requestType + ":" + name)
	if err != nil {
		d.logger.Error("Failed to get values", zap.Error(err))
		return
	}

	ttl := uint32(300)

	if requestType == "node" {
		ttl = 3600
	}

	for _, value := range values {
		txt := &dns.TXT{
			Hdr: dns.RR_Header{
				Name:   utils.ToLowerCase(q.Name),
				Rrtype: dns.TypeTXT,
				Class:  dns.ClassINET,
				Ttl:    ttl,
			},
			Txt: []string{value},
		}
		msg.Answer = append(msg.Answer, txt)
	}
}

func handleTxtAuthRequest(msg *dns.Msg, q dns.Question) {
	// Create the TXT record
	txt := &dns.TXT{
		Hdr: dns.RR_Header{
			Name:   utils.ToLowerCase(q.Name),
			Rrtype: dns.TypeTXT,
			Class:  dns.ClassINET,
			Ttl:    300,
		},
		Txt: []string{"asdasd"},
	}
	msg.Answer = append(msg.Answer, txt)

}

func GenerateClientCookie() []byte {
	// In a real-world scenario, this should be a securely generated value
	// For simplicity, we use a fixed value here
	return []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
}

func handleARequest(msg *dns.Msg, q dns.Question) {
	a := &dns.A{
		Hdr: dns.RR_Header{
			Name:   utils.ToLowerCase(q.Name),
			Rrtype: dns.TypeA,
			Class:  dns.ClassINET,
			Ttl:    300,
		},
		A: net.ParseIP("128.140.37.196"),
	}
	msg.Answer = append(msg.Answer, a)
}

func handleAAAARequest(msg *dns.Msg, q dns.Question) {
	aaaa := &dns.AAAA{
		Hdr: dns.RR_Header{
			Name:   utils.ToLowerCase(q.Name),
			Rrtype: dns.TypeAAAA,
			Class:  dns.ClassINET,
			Ttl:    300,
		},
		AAAA: net.ParseIP("2a01:4f8:1c0c:68c1::1"),
	}
	msg.Answer = append(msg.Answer, aaaa)
}

func handleNotImplemented(msg *dns.Msg) {
	msg.Rcode = dns.RcodeNotImplemented
}

func IsUDPRequest(addr net.Addr) bool {
	switch addr.(type) {
	case *net.UDPAddr:
		return true
	case *net.TCPAddr:
		return false
	default:
		return false
	}
}
