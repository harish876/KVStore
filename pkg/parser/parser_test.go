package parser_test

import (
	"encoding/base64"
	"fmt"
	"regexp"
	"testing"

	"github.com/codecrafters-io/redis-starter-go/pkg/parser"
	"github.com/stretchr/testify/assert"
)

func TestParserEcho(t *testing.T) {
	input := []byte("*2\r\n$4\r\necho\r\n$3\r\nhey\r\n")
	msg, _ := parser.Decode(input)
	fmt.Printf("\nMethod: %s\nMessage: %s\nMessage length: %d\nSegment Length: %d\n", msg.Method, msg.Messages, msg.MessagesLength, msg.SegmentLength)

	if msg.Method != "echo" {
		t.Fatalf("Test Failed. Expected Method: echo\nRecieved Method: %s\n", msg.Method)
	}
	if msg.MessagesLength > 1 && msg.Messages[0] != "hey" {
		t.Fatalf("Test Failed. Expected Messages: hey\nRecieved Method: %s\n", msg.Messages)
	}
	if msg.MessagesLength > 1 && len(msg.Messages[0]) != 3 {
		t.Fatalf("Test Failed. Expected Messagelength: 3\nRecieved Method: %d\n", msg.MessagesLength)
	}
}
func TestParserPing(t *testing.T) {
	input := []byte("*1\r\n$4\r\nping\r\n")
	msg, _ := parser.Decode(input)
	fmt.Printf("\nMethod: %s\nMessage: %s\nMessage length: %d\nSegment Length: %d\n", msg.Method, msg.Messages, msg.MessagesLength, msg.SegmentLength)

	if msg.Method != "ping" {
		t.Fatalf("Test Failed. Expected Method: ping\nRecieved Method: %s\n", msg.Method)
	}
	if len(msg.Messages) != 0 {
		t.Fatalf("Test Failed. Expected Messages: \nRecieved Method: %s\n", msg.Messages)
	}
	if msg.MessagesLength != 0 {
		t.Fatalf("Test Failed. Expected Messagelength: 0\nRecieved Method: %d\n", msg.MessagesLength)
	}
}

func TestParserSet(t *testing.T) {
	input := []byte("*3\r\n$3\r\nset\r\n$4\r\nfoo\r\nbar\r\n")
	msg, err := parser.Decode(input)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("\nMethod: %s\nMessage: %s\nMessage length: %d\nSegment Length: %d\n", msg.Method, msg.Messages, msg.MessagesLength, msg.SegmentLength)
}

func TestParserGet(t *testing.T) {
	input := []byte("*2\r\n$3\r\nget\r\n$3\r\nbar\r\n")
	msg, err := parser.Decode(input)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("\nMethod: %s\nMessage: %s\nMessage length: %d\nSegment Length: %d\n", msg.Method, msg.Messages, msg.MessagesLength, msg.SegmentLength)
}

func TestParserSetExpiry(t *testing.T) {
	input := []byte("*5\r\n$3\r\nset\r\n$4\r\nfoo\r\nbar\r\n$2\r\npx\r\n$2\r\n100\r\n")
	msg, err := parser.Decode(input)
	if err != nil {
		t.Fatalf("Test Expiry %v", err)
	}
	fmt.Printf("\nMethod: %s\nMessage: %s\nMessage length: %d\nSegment Length: %d\n", msg.Method, msg.Messages, msg.MessagesLength, msg.SegmentLength)
}

func TestParserInfo(t *testing.T) {
	requestMessage := []string{"info", "replication"}
	input := parser.EncodeRespArray(requestMessage)
	msg, err := parser.Decode([]byte(input))
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("\nMethod: %s\nMessage: %s\nMessage length: %d\nSegment Length: %d\n", msg.Method, msg.Messages, msg.MessagesLength, msg.SegmentLength)
}

func TestReplOffset(t *testing.T) {
	var patternMatchError error
	responseValue := parser.EncodeRespString(fmt.Sprintf("role:%s\r\nmaster_replid:%s\r\nmaster_repl_offset:%d", "master", "8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb", 0))
	if regexp.MustCompile("master_replid:([a-zA-Z0-9]+)").Match([]byte(responseValue)) {
		fmt.Printf("Found master_replid:xxx in response.")
	} else {
		patternMatchError = fmt.Errorf("Expected master_replid:xxx to be present in response. Got: %q\n", responseValue)
	}

	if regexp.MustCompile("master_repl_offset:0").Match([]byte(responseValue)) {
		fmt.Printf("Found master_reploffset:0 in response.")
	} else {
		patternMatchError = fmt.Errorf("Expected master_repl_offset:0 to be present in response. Got: %q\n", responseValue)
	}
	if patternMatchError != nil {
		t.Fatal(patternMatchError)
	}
}

func TestMasterPing(t *testing.T) {
	fmt.Printf("\nPING:c%s\nREPLCONF1: %s\nREPLCONF2: %s\n",
		parser.EncodeRespArray([]string{"PING"}),
		parser.EncodeRespArray([]string{"REPLCONF", "listening-port", fmt.Sprintf("%d", 6380)}),
		parser.EncodeRespArray([]string{"REPLCONF", "capa", "psync2"}),
	)

	input := parser.EncodeRespArray([]string{"REPLCONF", "listening-port", fmt.Sprintf("%d", 6380)})
	msg, err := parser.Decode([]byte(input))
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("\nMethod: %s\nMessage: %s\nMessage length: %d\nSegment Length: %d\n", msg.Method, msg.Messages, msg.MessagesLength, msg.SegmentLength)

}

func TestParserPsync(t *testing.T) {
	input := parser.EncodeRespArray([]string{"PSYNC", "?", fmt.Sprintf("%d", -1)})
	msg, err := parser.Decode([]byte(input))
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("\nMethod: %s\nMessage: %s\nMessage length: %d\nSegment Length: %d\n", msg.Method, msg.Messages, msg.MessagesLength, msg.SegmentLength)
}

func TestRdb(t *testing.T) {
	base64String := "UkVESVMwMDEx+glyZWRpcy12ZXIFNy4yLjD6CnJlZGlzLWJpdHPAQPoFY3RpbWXCbQi8ZfoIdXNlZC1tZW3CsMQQAPoIYW9mLWJhc2XAAP/wbjv+wP9aog=="
	response, err := base64.StdEncoding.DecodeString(base64String)
	if err != nil {
		t.Fatal("Error decoding Base64:", err)
		return
	}
	fmt.Println(len(response), string(response))
}

func TestParserEmpty(t *testing.T) {
	input := []byte("")
	msg, err := parser.Decode(input)
	if err != nil {
		t.Fatalf("Test Expiry %v", err)
	}
	fmt.Printf("\nMethod: %s\nMessage: %s\nMessage length: %d\nSegment Length: %d\n", msg.Method, msg.Messages, msg.MessagesLength, msg.SegmentLength)
}

func TestGetAck(t *testing.T) {
	input := []byte("*3\r\n$8\r\nreplconf\r\n$6\r\ngetack\r\n$1\r\n*\r\n")
	msg, err := parser.Decode(input)
	if err != nil {
		t.Fatalf("Replconf get ack parsing failed %v", err)
	}
	assert.Equal(t, msg.Method, "replconf")
	response := parser.EncodeRespArray([]string{"REPLCONF", "ACK", "0"})
	assert.Equal(t, response, "*3\r\n$8\r\nREPLCONF\r\n$3\r\nACK\r\n$1\r\n0\r\n")
	assert.ElementsMatch(t, msg.Messages, []string{"getack", "*"})
}

func TestSetMessage(t *testing.T) {

}
