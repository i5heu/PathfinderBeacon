package reqLogic

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"

	"github.com/i5heu/PathfinderBeacon/pkg/utils"
	"github.com/miekg/dns"
)

func (d *ReqLogic) AddValue(key string, value string, ttl int) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	var values []string
	existingValue, err := d.store.Get([]byte(key))
	if err == nil {
		if err := json.Unmarshal(existingValue, &values); err != nil {
			return err
		}
	}

	values = append(values, value)
	values = utils.RemoveDuplicate(values)

	data, err := json.Marshal(values)
	if err != nil {
		return err
	}

	return d.store.Set([]byte(key), data, ttl)
}

func (d *ReqLogic) GetValues(key string) ([]string, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	var values []string
	data, err := d.store.Get([]byte(key))
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(data, &values); err != nil {
		return nil, err
	}

	return values, nil
}

func rateLimit(ctx context.Context, w dns.ResponseWriter, d *ReqLogic) error {
	if IsUDPRequest(w.RemoteAddr()) {
		_, _, _, ok, err := d.rateLimitStoreUDP.Take(ctx, getIPFromRemoteAddr(w.RemoteAddr().String()))
		if err != nil {
			log.Printf("Failed to get rate limit: UDP %s\n", err)
			return err
		}
		if !ok {
			log.Printf("Rate limit exceeded for UDP %s\n", w.RemoteAddr().String())
			return fmt.Errorf("rate limit exceeded udp")
		}
		return nil
	} else {
		_, _, _, ok, err := d.rateLimitStoreUDP.Take(ctx, getIPFromRemoteAddr(w.RemoteAddr().String()))
		if err != nil {
			log.Printf("Failed to get rate limit TCP: %s\n", err)
			return err
		}
		if !ok {
			log.Printf("Rate limit exceeded for TCP %s\n", w.RemoteAddr().String())
			return fmt.Errorf("rate limit exceeded")
		}
		return nil
	}
}

func getIPFromRemoteAddr(remoteAddr string) string {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return remoteAddr
	}
	return host
}
