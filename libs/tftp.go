package libs

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	TFTPPort        = 69
	TFTPTimeout     = 180 * time.Second
)

// TFTPServer represents a TFTP server instance
type TFTPServer struct {
	conn   *net.UDPConn
	ctx    context.Context
	cancel context.CancelFunc
}

// NewTFTPServer creates a new TFTP server instance
func NewTFTPServer() *TFTPServer {
	ctx, cancel := context.WithCancel(context.Background())
	return &TFTPServer{
		ctx:    ctx,
		cancel: cancel,
	}
}

// Start begins listening for TFTP requests
func (t *TFTPServer) Start() error {
	// Create tftp directory
	os.MkdirAll("tftp", 0755)
	
	addr, err := net.ResolveUDPAddr("udp", ":69")
	if err != nil {
		return err
	}

	t.conn, err = net.ListenUDP("udp", addr)
	if err != nil {
		return err
	}
	defer t.conn.Close()

	fmt.Println("[INFO]: TFTP Server started on port 69")

	buffer := make([]byte, 65536)
	for {
		select {
		case <-t.ctx.Done():
			return nil
		default:
			t.conn.SetReadDeadline(time.Now().Add(1 * time.Second))
			n, clientAddr, err := t.conn.ReadFromUDP(buffer)
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue
				}
				fmt.Printf("TFTP read error: %v\n", err)
				continue
			}

			go t.handleRequest(buffer[:n], clientAddr)
		}
	}
}

// Stop stops the TFTP server
func (t *TFTPServer) Stop() {
	if t.cancel != nil {
		t.cancel()
	}
	if t.conn != nil {
		t.conn.Close()
	}
}

// handleRequest processes incoming TFTP requests
func (t *TFTPServer) handleRequest(data []byte, clientAddr *net.UDPAddr) {
	if len(data) < 4 {
		return
	}

	opcode := (int(data[0]) << 8) | int(data[1])
	
	switch opcode {
	case 1: // RRQ (Read Request)
		t.handleReadRequest(data[2:], clientAddr)
	case 2: // WRQ (Write Request)
		t.handleWriteRequest(data[2:], clientAddr)
	}
}

// handleReadRequest processes TFTP read requests (client wants to download file)
func (t *TFTPServer) handleReadRequest(data []byte, clientAddr *net.UDPAddr) {
	parts := strings.Split(string(data), "\x00")
	if len(parts) < 2 {
		return
	}
	
	filename := parts[0]
	mode := parts[1]
	
	fmt.Printf("[INFO]: TFTP GET request for %s from %s\n", filename, clientAddr.IP)
	
	filePath := filepath.Join("tftp", filename)
	file, err := os.Open(filePath)
	if err != nil {
		t.sendError(clientAddr, 1, "File not found")
		return
	}
	defer file.Close()
	
	t.sendFile(file, clientAddr, mode)
}

// handleWriteRequest processes TFTP write requests (client wants to upload file)
func (t *TFTPServer) handleWriteRequest(data []byte, clientAddr *net.UDPAddr) {
	parts := strings.Split(string(data), "\x00")
	if len(parts) < 2 {
		return
	}
	
	filename := parts[0]
	mode := parts[1]
	
	fmt.Printf("[INFO]: TFTP PUT request for %s from %s\n", filename, clientAddr.IP)
	
	filePath := filepath.Join("tftp", filename)
	file, err := os.Create(filePath)
	if err != nil {
		t.sendError(clientAddr, 2, "Cannot create file")
		return
	}
	defer file.Close()
	
	t.receiveFile(file, clientAddr, mode)
}

// sendFile sends a file to the TFTP client
func (t *TFTPServer) sendFile(file *os.File, clientAddr *net.UDPAddr, mode string) {
	blockNum := uint16(1)
	buffer := make([]byte, 512)
	
	for {
		n, err := file.Read(buffer)
		if err != nil && err != io.EOF {
			t.sendError(clientAddr, 0, "Read error")
			return
		}
		
		// Create DATA packet
		packet := make([]byte, 4+n)
		packet[0] = 0
		packet[1] = 3 // DATA opcode
		packet[2] = byte(blockNum >> 8)
		packet[3] = byte(blockNum & 0xFF)
		copy(packet[4:], buffer[:n])
		
		t.conn.WriteToUDP(packet, clientAddr)
		
		if n < 512 {
			break
		}
		
		blockNum++
	}
}

// receiveFile receives a file from the TFTP client
func (t *TFTPServer) receiveFile(file *os.File, clientAddr *net.UDPAddr, mode string) {
	// Send ACK 0
	ack := []byte{0, 4, 0, 0}
	t.conn.WriteToUDP(ack, clientAddr)
	
	expectedBlock := uint16(1)
	buffer := make([]byte, 65536)
	
	for {
		t.conn.SetReadDeadline(time.Now().Add(TFTPTimeout))
		n, addr, err := t.conn.ReadFromUDP(buffer)
		if err != nil {
			fmt.Printf("TFTP receive error: %v\n", err)
			return
		}
		
		if !addr.IP.Equal(clientAddr.IP) || addr.Port != clientAddr.Port {
			continue
		}
		
		if n < 4 {
			continue
		}
		
		opcode := (int(buffer[0]) << 8) | int(buffer[1])
		if opcode != 3 { // Not DATA
			continue
		}
		
		blockNum := (uint16(buffer[2]) << 8) | uint16(buffer[3])
		if blockNum != expectedBlock {
			continue
		}
		
		// Write data to file
		dataLen := n - 4
		if dataLen > 0 {
			file.Write(buffer[4:4+dataLen])
		}
		
		// Send ACK
		ack := []byte{0, 4, buffer[2], buffer[3]}
		t.conn.WriteToUDP(ack, clientAddr)
		
		if dataLen < 512 {
			fmt.Printf("[INFO]: File transfer completed for %s\n", clientAddr.IP)
			break
		}
		
		expectedBlock++
	}
}

// sendError sends a TFTP error packet to the client
func (t *TFTPServer) sendError(clientAddr *net.UDPAddr, errorCode int, message string) {
	packet := make([]byte, 4+len(message)+1)
	packet[0] = 0
	packet[1] = 5 // ERROR opcode
	packet[2] = byte(errorCode >> 8)
	packet[3] = byte(errorCode & 0xFF)
	copy(packet[4:], message)
	packet[len(packet)-1] = 0
	
	t.conn.WriteToUDP(packet, clientAddr)
}
