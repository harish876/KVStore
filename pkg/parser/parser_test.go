package parser_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/codecrafters-io/redis-starter-go/pkg/parser"
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
	responseMessage := parser.EncodeResponse([]string{
		parser.GetLablelledMessage("role", "master"),
		parser.GetLablelledMessage("master_replid", "8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb"),
		parser.GetLablelledMessage("master_repl_offset", 0),
	})
	input := parser.EncodeRequest(requestMessage)
	msg, err := parser.Decode([]byte(input))
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("\nMethod: %s\nMessage: %s\nMessage length: %d\nSegment Length: %d\n", msg.Method, msg.Messages, msg.MessagesLength, msg.SegmentLength)
	fmt.Printf("\nResponse Message is: %s", responseMessage)
}

func TestReplOffset(t *testing.T) {
	var patternMatchError error
	responseValue := parser.EncodeResponse([]string{fmt.Sprintf("role:%s\r\nmaster_replid:%s\r\nmaster_repl_offset:%d", "master", "8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb", 0)})
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
