package parser

import (
	"fmt"
	"strconv"
	"strings"
)

type RespMessage struct {
	Method        string
	Message       string
	MessageLength int
	MethodLength  int
	SegmentLength int //Number of segments - rename it later
}

// TODO: Improve parser
func Parse(msg []byte) RespMessage {
	strMsg := string(msg)
	splitMsg := strings.Split(strMsg, "\r\n")
	var parsedArray []string
	var parsedLenArray []int
	var parsedRespMessage RespMessage
	for _, v := range splitMsg {
		if strings.Contains(v, "$") {
			msgLenArray := strings.Split(v, "$")
			if len(msgLenArray) == 2 {
				length, err := strconv.Atoi(msgLenArray[1])
				if err != nil {
					fmt.Printf("Error parsing message length: %v", err)
				}
				parsedLenArray = append(parsedLenArray, length)
			}
			continue
		}
		if strings.Contains(v, "*") {
			segLenArray := strings.Split(v, "*")
			if len(segLenArray) == 2 {
				length, err := strconv.Atoi(segLenArray[1])
				if err != nil {
					fmt.Printf("Error parsing message length: %v", err)
				}
				parsedRespMessage.SegmentLength = length
			}
			continue
		}
		if v == "" {
			continue
		}
		parsedArray = append(parsedArray, v)
	}
	if len(parsedArray) == 1 {
		//This is a ping command, which doesnt contain any method
		parsedRespMessage.Method = "ping"
		parsedRespMessage.MethodLength = 4
		parsedRespMessage.Message = parsedArray[0]
	} else if len(parsedArray) >= 2 {
		parsedRespMessage.Method = strings.ToLower(parsedArray[0])
		parsedRespMessage.Message = parsedArray[1]
	}

	if len(parsedLenArray) == 1 {
		//There was no method in the request
		parsedRespMessage.MessageLength = parsedLenArray[0]
	} else if len(parsedArray) >= 2 {
		parsedRespMessage.MethodLength = parsedLenArray[0]
		parsedRespMessage.MessageLength = parsedLenArray[1]
	}
	return parsedRespMessage
}
