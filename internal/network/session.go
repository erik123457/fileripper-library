// This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
// If a copy of the MPL was not distributed with this file, You can obtain one at
// https://mozilla.org/MPL/2.0/.

package network

import (
	"fmt"
	"net"
	"strings"
	"time"

	"fileripper/internal/core"
)

// SftpSession holds the raw connection state.
type SftpSession struct {
	Hostname string
	Port     int
	conn     net.Conn 
}

func NewSession(host string, port int) *SftpSession {
	return &SftpSession{
		Hostname: host,
		Port:     port,
	}
}

// Connect opens a raw TCP connection.
func (s *SftpSession) Connect() error {
	address := fmt.Sprintf("%s:%d", s.Hostname, s.Port)
	fmt.Printf(">> Network: Connecting to %s...\n", address)

	conn, err := net.DialTimeout("tcp", address, 5*time.Second)
	if err != nil {
		fmt.Printf(">> Network: Connection failed: %v\n", err)
		return core.ErrConnectionFailed
	}
	s.conn = conn

	buffer := make([]byte, 128)
	s.conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	
	n, err := s.conn.Read(buffer)
	if err != nil {
		fmt.Printf(">> Network: Failed to read handshake: %v\n", err)
		return core.ErrConnectionFailed
	}

	banner := string(buffer[:n])
	banner = strings.TrimSpace(banner)

	fmt.Printf(">> Network: Handshake Success! Server said: '%s'\n", banner)

	if !strings.HasPrefix(banner, "SSH-") {
		fmt.Println(">> Network: WARNING. This doesn't look like an SSH server.")
	}

	return nil
}

func (s *SftpSession) Close() {
	if s.conn != nil {
		s.conn.Close()
	}
}