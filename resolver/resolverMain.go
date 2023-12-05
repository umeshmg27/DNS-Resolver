package resolver

import (
	"errors"
	"fmt"
	"net"
)

// Global variables to avoid passing variables around.
const udpMaxPacketSize = 512
const headerSize = 12

// DNSMessage - Complete message
type DNSMessage struct {
	Header     Header
	Questions  []Question
	Answers    []ResourceRecord
	Authority  []ResourceRecord
	Additional []ResourceRecord
}

type nameServers []string

// Push adds an item to the top of the stack
func (s *nameServers) Push(item string) {
	*s = append(*s, item)
}

// Pop removes the item from the top of the stack and returns it
func (s *nameServers) Pop() (string, error) {
	if len(*s) == 0 {
		return "", errors.New("pop from empty stack")
	}

	index := len(*s) - 1   // Get the index of the top most element.
	element := (*s)[index] // Index into the slice and obtain the element.
	*s = (*s)[:index]      // Remove it from the stack by slicing it off.
	return element, nil
}

func ConstructDnsMessage(domainName string, nameServer string) ([]byte, uint16, error) {
	header := Header{
		ID:                    generateRandomNumber(),
		Flags:                 256,
		QuestionCount:         1,
		AnswerRecordCount:     0,
		AuthorityRecordCount:  0,
		AdditionalRecordCount: 0,
	}
	question := Question{
		Name:  domainName,
		Type:  1,
		Class: 1,
	}
	headerEncoded := header.Encode()
	questionEncoded, err := question.EncodeQuestion()
	if err != nil {
		return nil, 0, err
	}
	QueryMessage := append(headerEncoded, questionEncoded...)
	return QueryMessage, header.ID, nil
}

func HandleDNSRequest(domainName string, nameServer string) ([]string, string, error) {

	queryMessage, reqID, err := ConstructDnsMessage(domainName, nameServer)
	if err != nil {
		return nil, "", err
	}
	visitedNS := make(map[string]bool)
	NSInQueue := nameServers{nameServer}

	for len(NSInQueue) > 0 {
		curNsIp, err := NSInQueue.Pop()
		if err != nil {
			return nil, "", err
		}

		conn, err := net.Dial("udp", fmt.Sprintf("%s:53", curNsIp))
		if err != nil {
			return nil, "", err
		}
		defer conn.Close()

		_, err = conn.Write(queryMessage)
		if err != nil {
			return nil, "", err
		}

		buffer := make([]byte, udpMaxPacketSize)
		_, err = conn.Read(buffer)
		if err != nil {
			return nil, "", err
		}

		bufferPosition := 0
		responseHeader, err := DecodeHeader(buffer)
		if err != nil {
			return nil, "", err
		}
		err = VerifyHeader(responseHeader, reqID)
		if err != nil {

			return nil, "", err
		}

		bufferPosition += headerSize
		responseBody, size, err := DecodeQuestion(buffer, bufferPosition)
		if err != nil || responseBody.Name != domainName {
			return nil, "", err
		}
		bufferPosition += size
		answerRecordData := []string{}
		for i := 0; i < int(responseHeader.AnswerRecordCount); i++ {
			answer, _, err := DecodeResource(buffer, bufferPosition)
			if err != nil {
				return nil, "", err
			}
			fmt.Printf("\n answer %+v \n", answer)
			answerRecordData = append(answerRecordData, fmt.Sprintf("%d.%d.%d.%d", answer.Data[0], answer.Data[1], answer.Data[2], answer.Data[3]))
		}
		if len(answerRecordData) > 0 {
			return answerRecordData, answerRecordData[0], nil
		}

		authorityRecords := make([]*ResourceRecord, 0)
		for i := 0; i < int(responseHeader.AuthorityRecordCount); i++ {
			authority, size, err := DecodeResource(buffer, bufferPosition)
			if err != nil {
				return nil, "", err
			}
			authorityRecords = append(authorityRecords, authority)
			bufferPosition += size
		}

		additionalRecords := make([]*ResourceRecord, 0)
		for i := 0; i < int(responseHeader.AdditionalRecordCount); i++ {
			additional, size, err := DecodeResource(buffer, bufferPosition)
			if err != nil {
				return nil, "", err
			}
			additionalRecords = append(additionalRecords, additional)
			bufferPosition += size
		}
		for i := range additionalRecords {
			// We have ipv4 address for server that can help resolve the query.
			ar := additionalRecords[i]
			if ar.Type == 1 && ar.Class == 1 && ar.rdLEngth == 4 {
				newIP := fmt.Sprintf("%d.%d.%d.%d", ar.Data[0], ar.Data[1], ar.Data[2], ar.Data[3])
				if _, exists := visitedNS[newIP]; !exists {
					NSInQueue.Push(newIP)
					visitedNS[newIP] = true
				}
			}
		}

		// Need to resolve name server's ip address to continue.
		if len(NSInQueue) == 0 && len(authorityRecords) > 0 {
			fmt.Println("Name Server IP.")
			_, nameServer, err := HandleDNSRequest(string(authorityRecords[0].Data), "8.8.8.8")
			if err != nil {
				return nil, "", err
			}
			NSInQueue.Push(nameServer)
		}

	}
	return nil, "", fmt.Errorf("failed to resolve this domain name.")

}
