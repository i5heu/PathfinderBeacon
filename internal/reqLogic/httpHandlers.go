package reqLogic

import (
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"

	"github.com/i5heu/PathfinderBeacon/pkg/auth"
	"github.com/i5heu/PathfinderBeacon/pkg/utils"
)

func validateAndParseRegisteringAddress(regString string) (utils.RegisteringNode, error) {
	if regString == "" {
		return utils.RegisteringNode{}, fmt.Errorf("address is empty")
	}

	var regAddr utils.RegisteringNode
	err := json.Unmarshal([]byte(regString), &regAddr)
	if err != nil {
		return utils.RegisteringNode{}, err
	}

	if regAddr.Room == "" {
		return utils.RegisteringNode{}, fmt.Errorf("room is empty")
	}
	if len(regAddr.Addresses) == 0 {
		return utils.RegisteringNode{}, fmt.Errorf("addresses are empty")
	}
	if len(regAddr.Addresses) > 50 {
		return utils.RegisteringNode{}, fmt.Errorf("too many addresses")
	}

	if !utils.CheckIfSha224(regAddr.Room) {
		return utils.RegisteringNode{}, fmt.Errorf("room is not a valid sha224 hash")
	}

	// check if ip is valid
	for _, addr := range regAddr.Addresses {
		if net.ParseIP(addr.Ip) == nil {
			return utils.RegisteringNode{}, fmt.Errorf("ip is not valid: %s", addr.Ip)
		}
	}

	// check if port is valid
	for _, addr := range regAddr.Addresses {
		if addr.Port < 1 || addr.Port > 65535 {
			return utils.RegisteringNode{}, fmt.Errorf("port is not valid %d", addr.Port)
		}
	}

	// check if protocol is valid
	for _, addr := range regAddr.Addresses {
		if addr.Protocol != "tcp" && addr.Protocol != "udp" {
			return utils.RegisteringNode{}, fmt.Errorf("protocol is not valid %s", addr.Protocol)
		}
	}

	return regAddr, nil
}

func (d *ReqLogic) RegisterNodeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		fmt.Println("Method not allowed")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	fmt.Println("Request received", r.Method, r.URL.Path)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Println("Failed to read body", err)
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}

	regNode, err := validateAndParseRegisteringAddress(string(body))
	if err != nil {
		fmt.Println("Failed to parse body", err)
		http.Error(w, fmt.Errorf("Failed to parse body: %s ", err).Error(), http.StatusBadRequest)
		return
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		fmt.Println("Failed to split host port", err)
		http.Error(w, "Failed to split host port", http.StatusInternalServerError)
		return
	}

	// if host is from docker container, get the real ip via X-Real-Ip header
	if strings.HasPrefix(host, "172.19.0.") && r.Header.Get("X-Real-Ip") != "" {
		host = r.Header.Get("X-Real-Ip")
	}

	// verify the roomName with the roomSignature
	ok, err := auth.VerifyRoomSignature(regNode.Room, regNode.RoomSignature, regNode.PublicKey)
	if err != nil {
		http.Error(w, fmt.Errorf("Failed to verify room signature %w", err).Error(), http.StatusBadRequest)
		return
	}
	if !ok {
		http.Error(w, "Failed to verify room signature", http.StatusUnauthorized)
		return
	}

	nodeName := sha512.Sum512_224([]byte("node:" + host))

	// set ttl to infinite if it is the demo room
	ttl := 3600
	if d.demoRoomName == regNode.Room {
		ttl = 0
	}

	err = d.AddValue("room:"+regNode.Room, hex.EncodeToString(nodeName[:]), ttl)
	if err != nil {
		fmt.Println("Failed to add value", err)
		http.Error(w, "Failed to add value", http.StatusInternalServerError)
		return
	}

	for _, addr := range regNode.Addresses {
		err = d.AddValue("node:"+hex.EncodeToString(nodeName[:]), fmt.Sprintf("%s://%s:%d", addr.Protocol, addr.Ip, addr.Port), ttl)
		if err != nil {
			fmt.Println("Failed to add value", err)
			http.Error(w, "Failed to add value", http.StatusInternalServerError)
			return
		}
	}

	fmt.Println("Node registered", hex.EncodeToString(nodeName[:]), "from IP", host)
	w.WriteHeader(http.StatusOK)
}

func (d *ReqLogic) Greet(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Request received", r.Method, r.URL.Path)
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Hello World!")
}

func (d *ReqLogic) LandingPage(w http.ResponseWriter, r *http.Request) {
	file := "template/index.html"
	http.ServeFile(w, r, file)
}
