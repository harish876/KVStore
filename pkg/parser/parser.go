package parser

import (
	"fmt"
	"strconv"
	"strings"
)

type RESPMessage struct {
	Method         string
	Messages       []string
	MessagesLength int
	SegmentLength  int //Number of segments - rename it later
}

// TODO: Improve parser logic
func Decode(msg []byte) (RESPMessage, error) {
	strMsg := string(msg)
	splitMsg := strings.Split(strMsg, "\r\n")

	var parsedArray []string
	var parsedLenArray []int
	var parsedRespMessage RESPMessage

	for _, v := range splitMsg {
		if strings.Contains(v, "*") {
			segLenArray := strings.Split(v, "*")
			if len(segLenArray) != 2 {
				return parsedRespMessage, fmt.Errorf("error parsing segment. segment length not provided")
			}
			length, err := strconv.Atoi(segLenArray[1])
			if err != nil {
				return parsedRespMessage, fmt.Errorf("error parsing segment length: %v", err)
			}
			parsedRespMessage.SegmentLength = length
			continue
		}

		if strings.Contains(v, "$") {
			msgLenArray := strings.Split(v, "$")
			if len(msgLenArray) == 2 {
				length, err := strconv.Atoi(msgLenArray[1])
				if err != nil {
					return parsedRespMessage, fmt.Errorf("error parsing message length: %v", err)
				}
				parsedLenArray = append(parsedLenArray, length)
			}
			continue
		}

		if v == "" {
			//END OF MESSAGE
			continue
		}
		parsedArray = append(parsedArray, v)
	}

	if len(parsedLenArray) == 0 {
		return parsedRespMessage, nil
	} else if len(parsedArray) == 1 {
		parsedRespMessage.Method = strings.ToLower(parsedArray[0])
	} else if len(parsedArray) >= 2 {
		parsedRespMessage.Method = strings.ToLower(parsedArray[0])
		parsedRespMessage.Messages = parsedArray[1:]
	}

	if parsedRespMessage.SegmentLength > 0 && parsedRespMessage.SegmentLength-1 != len(parsedRespMessage.Messages) {
		return parsedRespMessage, fmt.Errorf("unable to parse message. incorrect segment length or number of messages")
	}
	parsedRespMessage.MessagesLength = len(parsedRespMessage.Messages)
	return parsedRespMessage, nil
}

// TODO: make a encode function to make life easier
func Encode(msg RESPMessage) ([]byte, error) {
	return nil, nil
}
