package utils

import (
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"os"
	"time"
)

func GenerateReplicationId(default_host string) string {
	timestamp := time.Now().Unix()
	machineID, err := os.Hostname()
	if err != nil {
		machineID = default_host
	}
	data := fmt.Sprintf("%d%s", timestamp, machineID)
	hash := sha512.Sum512([]byte(data))
	hashHex := hex.EncodeToString(hash[:])
	truncatedHash := hashHex[:40]

	return truncatedHash
}
