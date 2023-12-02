package resolver

import (
	"fmt"
	"net"
)

// DNSMessage - Complete message
type DNSMessage struct {
	Header     Header
	Questions  []Question
	Answers    []ResourceRecord
	Authority  []ResourceRecord
	Additional []ResourceRecord
}

func HandleDNSRequest(conn *net.UDPConn) error {
	buffer := make([]byte, 512) // Max size for a DNS message
	n, addr, err := conn.ReadFromUDP(buffer)
	if err != nil {
		return fmt.Errorf("failed to read from UDP: %v", err)
	}
	fmt.Printf("\n Tadaaaaaaa %+v \n", buffer)

	// Decode the header
	header, err := DecodeHeader(buffer[:12])
	if err != nil {
		return fmt.Errorf("failed to decode header: %v", err)
	}

	fmt.Printf("\n header %+v \n", header)

	// Assume one question and start decoding it at byte 12
	question, err := DecodeQuestion(buffer[12:n])
	if err != nil {
		return fmt.Errorf("failed to decode question: %v", err)
	}
	fmt.Printf("\n DecodeQuestion %+v \n", question)

	// Dummy resolution - returning a fixed IP for any domain
	resolvedIP := "93.184.216.34" // Example IP address

	// Construct the response
	response, err := constructResponse(header, question, resolvedIP)
	if err != nil {
		return fmt.Errorf("failed to construct response: %v", err)
	}

	fmt.Printf("\nresponse %+v \n", string(response))

	// Send the response back to the client
	_, err = conn.WriteToUDP(response, addr)
	if err != nil {
		return fmt.Errorf("failed to write to UDP: %v", err)
	}

	return nil
}

func constructResponse(header *Header, question *Question, ip string) ([]byte, error) {
	// Modify the header for the response
	header.Flags = 0x8000 // Set response flag
	header.AnswerRecordCount = 1

	// Encode the header
	headerBuffer := header.Encode()

	// Re-encode the question
	questionBuffer, err := question.EncodeQuestion()
	if err != nil {
		return nil, err
	}

	// Construct the answer section
	answer := ResourceRecord{
		Name:  question.Name,
		Type:  TypeA,
		Class: ClassIN,
		TTL:   300, // Example TTL
		Data:  ip,
	}
	answerBuffer, err := answer.Encode()
	if err != nil {
		return nil, err
	}

	// Combine the header, question, and answer into the final response
	response := append(headerBuffer, questionBuffer...)
	response = append(response, answerBuffer...)

	return response, nil
}

func forwardQuery(buffer []byte) ([]byte, error) {
	// Forward the query to an external DNS server
	// This is a simplified example
	serverAddr := "8.8.8.8:53"
	conn, err := net.Dial("udp", serverAddr)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	_, err = conn.Write(buffer)
	if err != nil {
		return nil, err
	}

	response := make([]byte, 512)
	_, err = conn.Read(response)
	if err != nil {
		return nil, err
	}

	return response, nil
}
