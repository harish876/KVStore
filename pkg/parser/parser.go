package parser

import (
	"bufio"
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"
)

var (
	BULK_NULL_STRING = "$-1\r\n"
)

type RESPMessage struct {
	Method         string
	Messages       []string
	MessagesLength int
	SegmentLength  int
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
			if segLenArray[0] == "" && segLenArray[1] == "" {
				parsedArray = append(parsedArray, v)
				continue
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

/* Only works for array of length 1*/
func EncodeRespString(msg string) string {
	var result string
	result += fmt.Sprintf("$%d\r\n%s\r\n", len(msg), msg)
	return result
}

func EncodeRespArray(messages []string) string {
	if len(messages) > 0 {
		var result string
		result += fmt.Sprintf("*%d\r\n", len(messages))
		for _, msg := range messages {
			result += fmt.Sprintf("$%d\r\n%s\r\n", len(msg), msg)
		}
		return result
	} else {
		return BULK_NULL_STRING
	}
}

func EncodeSimpleString(msg string) string {
	return fmt.Sprintf("+%s\r\n", msg)
}

func GetLablelledMessage(label string, value any) string {
	if reflect.TypeOf(value).Kind() == reflect.Int {
		return fmt.Sprintf("%s:%d", label, value)
	} else if reflect.TypeOf(value).Kind() == reflect.String {
		return fmt.Sprintf("%s:%s", label, value)
	} else {
		return fmt.Sprintf("%s:%v", label, value)
	}
}

// TODO: Implement this later to have a nice message builder
func EncodeSingleMessage(messages []string) string {
	return ""
}

// Supports only +ve numbers for now
// :[<+|->]<value>\r\n
func EncodeInt(val int) string {
	return fmt.Sprintf(":%d\r\n", val)
}

func DecodeV1(reader *bufio.Reader) (arr []string, bytesRead int, err error) {
	var arrSize, strSize int
	for {
		var token string
		token, err = reader.ReadString('\n')
		if err != nil {
			return
		}
		bytesRead += len(token)
		token = strings.TrimRight(token, "\r\n")
		switch {
		case arrSize == 0 && token[0] == '*':
			arrSize, err = strconv.Atoi(token[1:])
			if err != nil {
				return
			}
		case strSize == 0 && token[0] == '$':
			strSize, err = strconv.Atoi(token[1:])
			if err != nil {
				return
			}
		default:
			if len(token) != strSize {
				log.Printf("[from master] Wrong string size - got: %d, want: %d\n", len(token), strSize)
				break
			}
			arrSize--
			strSize = 0
			arr = append(arr, token)
		}
		if arrSize == 0 {
			break
		}
	}
	return
}
