package resolver

import (
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

func HandleDNSRequest(domainName string, answerRecordData []string, nameServer string) ([]string, string, error) {

	queryMessage, reqID, err := ConstructDnsMessage(domainName, nameServer)
	if err != nil {
		return nil, "", err
	}
	visitedNS := make(map[string]bool)
	NSInQueue := nameServers{nameServer}

	for len(NSInQueue) > 0 {
		curNsIp := NSInQueue[0]
		NSInQueue = NSInQueue[1:]
		fmt.Printf("\n Querying for %+s in %s", domainName, curNsIp)
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

		for i := 0; i < int(responseHeader.AnswerRecordCount); i++ {
			answer, _, err := DecodeResource(buffer, bufferPosition)
			if err != nil {
				return nil, "", err
			}
			answerRecordData = append(answerRecordData, fmt.Sprintf("%d.%d.%d.%d", answer.Data[0], answer.Data[1], answer.Data[2], answer.Data[3]))
		}
		if len(answerRecordData) > 0 {
			return answerRecordData, "", nil
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
			// Get ip4 address for the query
			ar := additionalRecords[i]
			if ar.Type == 1 && ar.Class == 1 && ar.rdLEngth == 4 {
				newIP := fmt.Sprintf("%d.%d.%d.%d", ar.Data[0], ar.Data[1], ar.Data[2], ar.Data[3])
				if _, exists := visitedNS[newIP]; !exists {
					NSInQueue = append(NSInQueue, newIP)
					visitedNS[newIP] = true
				}
			}
		}

		// Resolve the NS data
		if len(NSInQueue) == 0 && len(authorityRecords) > 0 {
			_, nameServer, err := HandleDNSRequest(fmt.Sprintf("%d.%d.%d.%d", authorityRecords[0].Data[0], authorityRecords[0].Data[1], authorityRecords[0].Data[2], authorityRecords[0].Data[3]), answerRecordData, "192.168.6.181")
			if err != nil {
				return nil, "", err
			}
			NSInQueue = append(NSInQueue, nameServer)
		}

	}
	if len(answerRecordData) > 0 {
		return answerRecordData, answerRecordData[0], nil
	}
	return nil, "", fmt.Errorf("failed to resolve this domain name")

}
