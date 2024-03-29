package parser_test

import (
	"fmt"
	"testing"

	"github.com/codecrafters-io/redis-starter-go/pkg/parser"
)

func TestParserEcho(t *testing.T) {
	input := []byte("*2\r\n$4\r\necho\r\n$3\r\nhey\r\n")
	msg := parser.Parse(input)
	fmt.Printf("Method: %s\nMessage: %s\nMessage length: %d\nMethod Length: %d\nSegment Length: %d\n", msg.Method, msg.Message, msg.MessageLength, msg.MethodLength, msg.SegmentLength)

	if msg.Method != "echo" {
		t.Fatalf("Test Failed. Expected Method: echo\nRecieved Method: %s\n", msg.Method)
	}
	if msg.Message != "hey" {
		t.Fatalf("Test Failed. Expected Message: hey\nRecieved Method: %s\n", msg.Message)
	}
	if msg.MessageLength != 3 {
		t.Fatalf("Test Failed. Expected Messagelength: 3\nRecieved Method: %d\n", msg.MessageLength)
	}
	if msg.MethodLength != 4 {
		t.Fatalf("Test Failed. Expected MethodLength: 4\nRecieved Method: %d\n", msg.MethodLength)
	}
}
func TestParserPing(t *testing.T) {
	input := []byte("*1\r\n$4\r\nping\r\n")
	msg := parser.Parse(input)
	fmt.Printf("Method: %s\nMessage: %s\nMessage length: %d\nMethod Length: %d\nSegment Length: %d\n", msg.Method, msg.Message, msg.MessageLength, msg.MethodLength, msg.SegmentLength)

	if msg.Method != "ping" {
		t.Fatalf("Test Failed. Expected Method: ping\nRecieved Method: %s\n", msg.Method)
	}
	if msg.Message != "ping" {
		t.Fatalf("Test Failed. Expected Message: \nRecieved Method: %s\n", msg.Message)
	}
	if msg.MessageLength != 4 {
		t.Fatalf("Test Failed. Expected Messagelength: 0\nRecieved Method: %d\n", msg.MessageLength)
	}
	if msg.MethodLength != 4 {
		t.Fatalf("Test Failed. Expected MethodLength: 4\nRecieved Method: %d\n", msg.MethodLength)
	}
}
