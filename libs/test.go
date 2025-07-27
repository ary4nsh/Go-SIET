package libs

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	SmartInstallPort = 4786
	DefaultTimeout   = 5 * time.Second
)

// TestMode handles the vulnerability testing functionality
func (s *SIETClient) TestMode() {
	if s.Config.IPList != "" {
		s.testMultipleDevices()
	} else if s.Config.IP != "" {
		s.testSingleDevice(s.Config.IP)
	} else {
		fmt.Println("Error: No IP or IP list specified")
		os.Exit(1)
	}
}

// testMultipleDevices tests multiple devices from a file
func (s *SIETClient) testMultipleDevices() {
	file, err := os.Open(s.Config.IPList)
	if err != nil {
		fmt.Printf("Error opening IP list file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		ip := strings.TrimSpace(scanner.Text())
		if ip != "" {
			s.testSingleDevice(ip)
		}
	}
}

// testSingleDevice tests a single device for Smart Install vulnerability
func (s *SIETClient) testSingleDevice(ip string) {
	// Smart Install test packet
	testPacket := "00000001000000010000000400000008000000010000000000000000"
	expectedResponse := "00000004000000000000000300000008000000010000000000000000"
	
	testData, err := hex.DecodeString(testPacket)
	if err != nil {
		fmt.Printf("Error decoding test packet: %v\n", err)
		return
	}
	
	expectedData, err := hex.DecodeString(expectedResponse)
	if err != nil {
		fmt.Printf("Error decoding expected response: %v\n", err)
		return
	}

	conn, err := net.DialTimeout("tcp", net.JoinHostPort(ip, strconv.Itoa(SmartInstallPort)), DefaultTimeout)
	if err != nil {
		fmt.Printf("[ERROR]: Couldn't connect to %s: %v\n", ip, err)
		return
	}
	defer conn.Close()

	conn.SetDeadline(time.Now().Add(DefaultTimeout))
	
	_, err = conn.Write(testData)
	if err != nil {
		fmt.Printf("[ERROR]: Failed to send test packet to %s: %v\n", ip, err)
		return
	}

	response := make([]byte, len(expectedData))
	_, err = io.ReadAtLeast(conn, response, len(expectedData))
	if err != nil {
		fmt.Printf("[INFO]: Smart Install Director feature active on %s\n", ip)
		fmt.Printf("[INFO]: %s is not affected\n", ip)
		return
	}

	if string(response) == string(expectedData) {
		fmt.Printf("[INFO]: Smart Install Client feature active on %s\n", ip)
		fmt.Printf("[INFO]: %s is VULNERABLE\n", ip)
	} else {
		fmt.Printf("[ERROR]: Unexpected response from %s\n", ip)
		fmt.Printf("[INFO]: Unclear whether %s is affected\n", ip)
	}
}
