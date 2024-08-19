package main

import (
	// "crypto/tls"

	"html/template"
	"log"
	"net/http"
	"os"

	// "os"
	"time"

	"github.com/i5heu/PathfinderBeacon/internal/logg"
	"github.com/i5heu/PathfinderBeacon/internal/reqLogic"
	"github.com/i5heu/PathfinderBeacon/pkg/cache"
	"github.com/i5heu/PathfinderBeacon/pkg/rate_limiter"
	"go.uber.org/zap"
	// "golang.org/x/crypto/acme/autocert"
)

var logger *zap.Logger

func main() {
	logger = logg.InitLogger()
	defer logger.Sync()

	// get env example room name
	demoRoomName := os.Getenv("DEMO_ROOM_NAME")
	if demoRoomName == "" {
		demoRoomName = "NoDemoRoomName"
	}

	cacheStore := cache.NewCache(1000 * 1024 * 1024)
	defer cacheStore.Ticker.Stop()

	rateLimitStoreUDP, err := rate_limiter.NewRateLimiter(20, time.Minute*1)
	if err != nil {
		log.Fatal(err)
	}

	globalRateLimitStoreUDP, err := rate_limiter.NewRateLimiter(300, time.Minute*1)
	if err != nil {
		log.Fatal(err)
	}

	rateLimitStoreTCP, err := rate_limiter.NewRateLimiter(500, time.Minute*5)
	if err != nil {
		log.Fatal(err)
	}

	tmpl, err := template.ParseFiles("template/index.tmpl")
	if err != nil {
		log.Fatal(err)
		return
	}

	handler := reqLogic.NewDNSHandler(rateLimitStoreTCP, rateLimitStoreUDP, globalRateLimitStoreUDP, cacheStore, logger, demoRoomName, tmpl)

	go func() {
		reqLogic.StartDnsUdpServer(handler)
	}()

	go func() {
		reqLogic.StartDnsTcpServer(handler)
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("/register", handler.RegisterNodeHandler)
	mux.HandleFunc("/", handler.LandingPage)

	port := ":8088"
	if IsProductionMode() {
		port = ":80"
	}

	httpServer := &http.Server{
		Addr:    port,
		Handler: mux,
	}

	log.Println("Starting HTTP server on port ", port, "...")
	err = httpServer.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatalf("Failed to start HTTP server: %s\n", err)
	}

	if err != nil {
		log.Fatalf("Failed to start HTTPS server: %s\n", err)
	}
}

func IsProductionMode() bool {
	prod := os.Getenv("PROD_MODE")
	return (prod == "true")
}
