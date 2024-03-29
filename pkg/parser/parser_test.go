package parser_test

import (
	"fmt"
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
