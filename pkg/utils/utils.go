package utils

import (
	"strings"
)

type RegisteringAddress struct {
	Protocol string `json:"protocol"`
	Ip       string `json:"ip"`
	Port     int    `json:"port"`
}

type RegisteringNode struct {
	Room          string               `json:"room"`
	RoomSignature string               `json:"roomSignature"` // base64 encoded
	PublicKey     string               `json:"publicKey"`     // base64 encoded
	Addresses     []RegisteringAddress `json:"addresses"`
}

func ToLowerCase(s string) string {
	return strings.ToLower(s)
}

func RemoveDuplicate[T comparable](sliceList []T) []T {
	allKeys := make(map[T]bool)
	list := []T{}
	for _, item := range sliceList {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}

func CheckIfSha224(name string) bool {
	if len(name) != 56 {
		return false
	}

	for _, c := range name {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			return false
		}
	}

	return true
}

func GetDNSParameterAndCheckIfSha224(qName string) (string, bool) {
	// get the first part of the domain name
	parts := strings.Split(qName, ".")
	if len(parts) != 5 {
		return "", false
	}
	name := parts[0]

	if !CheckIfSha224(name) {
		return "", false
	}

	return name, true
}
