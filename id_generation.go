package main

import (
	"hash/fnv"
)

// Generate hash
func hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

// GetServerId - get server ID as hash
func GetServerId(host string, topic string) int {
	return int(hash(host + ":" + topic))
}
