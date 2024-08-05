package reqLogic

import (
	"log"
	"os"
	"sync"

	"github.com/i5heu/PathfinderBeacon/pkg/cache"
	"github.com/miekg/dns"
	"github.com/sethvargo/go-limiter"
	"go.uber.org/zap"
)

type ReqLogic struct {
	rateLimitStoreTCP       limiter.Store
	rateLimitStoreUDP       limiter.Store
	globalRateLimitStoreUDP limiter.Store
	mu                      sync.RWMutex
	store                   *cache.Cache
	logger                  *zap.Logger
}

func NewDNSHandler(rateLimitStoreTCP, rateLimitStore, globalRateLimitStore limiter.Store, store *cache.Cache, logger *zap.Logger) *ReqLogic {
	return &ReqLogic{
		rateLimitStoreTCP:       rateLimitStoreTCP,
		rateLimitStoreUDP:       rateLimitStore,
		globalRateLimitStoreUDP: globalRateLimitStore,
		store:                   store,
		logger:                  logger,
	}
}

func StartDnsUdpServer(handler *ReqLogic) {
	port := ":8053"
	if IsProductionMode() {
		port = ":53"
	}

	serverUDP := &dns.Server{Addr: port, Net: "udp"}
	defer serverUDP.Shutdown()

	dns.HandleFunc(".", handler.DNSReq)

	log.Println("Starting DNS UDP server on port ", port, "...")
	err := serverUDP.ListenAndServe()
	if err != nil {
		log.Fatalf("Failed to start DNS server: %s\n", err)
	}
}

func StartDnsTcpServer(handler *ReqLogic) {
	port := ":8053"
	if IsProductionMode() {
		port = ":53"
	}

	serverUDP := &dns.Server{Addr: port, Net: "tcp"}
	defer serverUDP.Shutdown()

	dns.HandleFunc(".", handler.DNSReq)

	log.Println("Starting DNS TCP server on port ", port, "...")
	err := serverUDP.ListenAndServe()
	if err != nil {
		log.Fatalf("Failed to start DNS server: %s\n", err)
	}
}

func IsProductionMode() bool {
	prod := os.Getenv("PROD_MODE")
	return (prod == "true")
}
