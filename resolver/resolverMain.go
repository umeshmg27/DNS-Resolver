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
	fmt.Printf("\n\n header %+v question %+v", header, question)
	headerEncoded := header.Encode()
	questionEncoded, err := question.EncodeQuestion()
	if err != nil {
		return nil, 0, err
	}
	fmt.Printf("\n\n header %+v question %+v", headerEncoded, questionEncoded)

	QueryMessage := append(headerEncoded, questionEncoded...)
	return QueryMessage, header.ID, nil
}

func HandleDNSRequest(domainName string, nameServer string) (string, error) {

	queryMessage, reqID, err := ConstructDnsMessage(domainName, nameServer)
	if err != nil {
		return "", err
	}

	fmt.Printf("\n\n query Message %+v", queryMessage)

	visitedNS := make(map[string]bool)
	NSInQueue := nameServers{nameServer}

	for len(NSInQueue) > 0 {
		curNsIp, err := NSInQueue.Pop()
		if err != nil {
			return "", err
		}

		conn, err := net.Dial("udp", fmt.Sprintf("%s:53", curNsIp))
		if err != nil {
			return "", err
		}
		defer conn.Close()

		_, err = conn.Write(queryMessage)
		if err != nil {
			return "", err
		}

		buffer := make([]byte, udpMaxPacketSize)
		_, err = conn.Read(buffer)
		if err != nil {
			return "", err
		}

		bufferPosition := 0
		responseHeader, err := DecodeHeader(buffer)
		if err != nil {
			return "", err
		}
		fmt.Printf("\n\n responseHeader %+b", responseHeader)
		err = VerifyHeader(responseHeader, reqID)
		if err != nil {

			return "", err
		}
		fmt.Printf("\n\n size %+v", headerSize)
		bufferPosition += headerSize
		responseBody, err := DecodeQuestion(buffer[12:])
		if err != nil || responseBody.Name != domainName+"." {
			fmt.Printf("\n\n responseBody %+v domain %+v", responseBody.Name[:], domainName)
			return "", err
		}
		bufferPosition += len(responseBody.Name)
		fmt.Printf("\n\n responseBody %+v", responseBody)
		for i := 0; i < int(responseHeader.AnswerRecordCount); i++ {
			answer, _, err := decodeResource(buffer, bufferPosition)
			if err != nil {
				return "", err
			}
			fmt.Printf("\n\n anserData %+v", answer)
			return fmt.Sprintf("%d.%d.%d.%d", answer.Data[0], answer.Data[1], answer.Data[2], answer.Data[3]), nil
		}

		authorityRecords := make([]*ResourceRecord, 0)
		for i := 0; i < int(responseHeader.AuthorityRecordCount); i++ {
			authority, size, err := decodeResource(buffer, bufferPosition)
			if err != nil {
				return "", err
			}
			authorityRecords = append(authorityRecords, authority)
			bufferPosition += size
		}

		additionalRecords := make([]*ResourceRecord, 0)
		for i := 0; i < int(responseHeader.AdditionalRecordCount); i++ {
			additional, size, err := decodeResource(buffer, bufferPosition)
			if err != nil {
				return "", err
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
			fmt.Println("Querying for name server ip.")
			nameServer, err := HandleDNSRequest(string(authorityRecords[0].Data), "8.8.8.8")
			if err != nil {
				return "", err
			}
			NSInQueue.Push(nameServer)
		}

	}
	return "", fmt.Errorf("failed to resolve this domain name.")

}

// func ProcessDNSRequest(conn *net.UDPConn) error {
// 	buffer := make([]byte, 512) // Max size for a DNS message
// 	n, addr, err := conn.ReadFromUDP(buffer)
// 	if err != nil {
// 		return fmt.Errorf("failed to read from UDP: %v", err)
// 	}
// 	fmt.Printf("\n Tadaaaaaaa %+v \n", buffer)

// 	// Decode the header
// 	header, err := DecodeHeader(buffer[:12])
// 	if err != nil {
// 		return fmt.Errorf("failed to decode header: %v", err)
// 	}

// 	fmt.Printf("\n header %+v \n", header)

// 	// Assume one question and start decoding it at byte 12
// 	question, err := DecodeQuestion(buffer[12:n])
// 	if err != nil {
// 		return fmt.Errorf("failed to decode question: %v", err)
// 	}
// 	fmt.Printf("\n DecodeQuestion %+v \n", question)

// 	// Dummy resolution - returning a fixed IP for any domain
// 	resolvedIP := "93.184.216.34" // Example IP address

// 	// Construct the response
// 	response, err := constructResponse(header, question, resolvedIP)
// 	if err != nil {
// 		return fmt.Errorf("failed to construct response: %v", err)
// 	}

// 	fmt.Printf("\nresponse %+v \n", response)

// 	// Send the response back to the client
// 	_, err = conn.WriteToUDP(response, addr)
// 	if err != nil {
// 		return fmt.Errorf("failed to write to UDP: %v", err)
// 	}

// 	return nil
// }

// func constructResponse(header *Header, question *Question, ip string) ([]byte, error) {
// 	// Modify the header for the response
// 	header.Flags = 0x8000 // Set response flag
// 	header.AnswerRecordCount = 1

// 	// Encode the header
// 	headerBuffer := header.Encode()

// 	// Re-encode the question
// 	questionBuffer, err := question.EncodeQuestion()
// 	if err != nil {
// 		return nil, err
// 	}

// 	// Construct the answer section
// 	answer := ResourceRecord{
// 		Name:  question.Name,
// 		Type:  TypeA,
// 		Class: ClassIN,
// 		TTL:   300, // Example TTL
// 		Data:  ip,
// 	}
// 	answerBuffer, err := answer.Encode()
// 	if err != nil {
// 		return nil, err
// 	}

// 	// Combine the header, question, and answer into the final response
// 	response := append(headerBuffer, questionBuffer...)
// 	response = append(response, answerBuffer...)

// 	return response, nil
// }
