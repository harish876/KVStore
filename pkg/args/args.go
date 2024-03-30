package args

import (
	"crypto/sha512"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type ReplicationConfig struct {
	ReplicationId     string
	ReplicationOffset int
}

type RedisArgs struct {
	ServerPort        int
	MasterHost        string
	MasterPort        int
	Role              string
	ReplicationConfig ReplicationConfig
}

var (
	DEFAULT_PORT = 6379
	DEFAULT_HOST = "localhost"
	MASTER_ROLE  = "master"
	SLAVE_ROLE   = "slave"
)

func ParseArgs() RedisArgs {
	currentPortPtr := flag.Int("port", DEFAULT_PORT, "Current Redis Server Port")
	masterServerDetailsPtr := flag.String("replicaof", "localhost 6379", "Current Redis Server Port")
	flag.Parse()
	port := *currentPortPtr
	masterServerDetails := *masterServerDetailsPtr
	var masterDetails = strings.Split(masterServerDetails, " ")

	var masterHost string
	var masterPort int
	var role string

	if len(masterDetails) < 2 {
		masterHost = DEFAULT_HOST
		masterPort = DEFAULT_PORT
	} else {
		parsedMasterPort, err := strconv.Atoi(masterDetails[1])
		if err != nil {
			masterPort = DEFAULT_PORT
		} else {
			masterPort = parsedMasterPort
		}
		masterHost = masterDetails[0]
	}

	if port == masterPort {
		role = MASTER_ROLE
	} else {
		role = SLAVE_ROLE
	}

	return RedisArgs{
		ServerPort: port,
		MasterPort: masterPort,
		MasterHost: masterHost,
		Role:       role,
		ReplicationConfig: ReplicationConfig{
			ReplicationId:     GenerateReplicationId(),
			ReplicationOffset: 0,
		},
	}
}

func GenerateReplicationId() string {
	timestamp := time.Now().Unix()
	machineID, err := os.Hostname()
	if err != nil {
		machineID = DEFAULT_HOST
	}
	data := fmt.Sprintf("%d%s", timestamp, machineID)
	hash := sha512.Sum512([]byte(data))
	hashHex := hex.EncodeToString(hash[:])
	truncatedHash := hashHex[:40]

	return truncatedHash
}
