package libs

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config represents the configuration for the SIET client
type Config struct {
	IP         string
	IPList     string
	Mode       string
	PublicIP   string
	ConfigFile string
	Username   string
	Password   string
	ReloadTime string
}

// SIETClient represents the Smart Install Exploitation Tool client
type SIETClient struct {
	Config *Config
	tftp   *TFTPServer
}

// NewSIETClient creates a new SIET client instance
func NewSIETClient(config *Config) *SIETClient {
	return &SIETClient{
		Config: config,
	}
}

// StartTFTPServer starts the TFTP server
func (s *SIETClient) StartTFTPServer() {
	s.tftp = NewTFTPServer()
	go s.tftp.Start()
	time.Sleep(1 * time.Second) // Give server time to start
}

// StopTFTPServer stops the TFTP server
func (s *SIETClient) StopTFTPServer() {
	if s.tftp != nil {
		s.tftp.Stop()
	}
}

// processMultipleIPs processes multiple IP addresses from a file
func (s *SIETClient) processMultipleIPs(handler func(string)) {
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
			handler(ip)
		}
	}
}

// SendPacketToDevice sends a packet to the specified device
func (s *SIETClient) SendPacketToDevice(ip string, packet []byte) error {
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(ip, strconv.Itoa(SmartInstallPort)), DefaultTimeout)
	if err != nil {
		return fmt.Errorf("connection failed: %v", err)
	}
	defer conn.Close()

	conn.SetDeadline(time.Now().Add(DefaultTimeout))
	
	_, err = conn.Write(packet)
	if err != nil {
		return fmt.Errorf("send failed: %v", err)
	}

	return nil
}

// GetLocalIP gets the local IP address for communicating with target IP
func (s *SIETClient) GetLocalIP(targetIP string) (string, error) {
	conn, err := net.Dial("udp", net.JoinHostPort(targetIP, "80"))
	if err != nil {
		return "", err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String(), nil
}

// CopyFile copies a file from source to destination
func (s *SIETClient) CopyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

// padCommand pads a command string to the specified length
func (s *SIETClient) padCommand(cmd string, length int) []byte {
	cmdBytes := []byte(cmd)
	if len(cmdBytes) >= length {
		return cmdBytes[:length]
	}
	
	padded := make([]byte, length)
	copy(padded, cmdBytes)
	return padded
}
