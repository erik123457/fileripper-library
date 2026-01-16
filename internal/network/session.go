/*
 * Copyright 2026 The FileRipper Team
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package network

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net"
	"time"

	"fileripper/internal/core"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// SftpSession holds the SSH connection state and the SFTP subsystem.
type SftpSession struct {
	Hostname   string
	Port       int
	User       string
	Password   string
	SshClient  *ssh.Client  // The tunnel
	SftpClient *sftp.Client // The file protocol wrapper
}

func NewSession(host string, port int, user, password string) *SftpSession {
	return &SftpSession{
		Hostname: host,
		Port:     port,
		User:     user,
		Password: password,
	}
}

// Connect establishes the secure SSH tunnel.
func (s *SftpSession) Connect() error {
	address := fmt.Sprintf("%s:%d", s.Hostname, s.Port)
	fmt.Printf(">> Network: Initiating Secure Handshake with %s...\n", address)

	authMethods := []ssh.AuthMethod{
		ssh.Password(s.Password),
	}

	hostKeyCallback := func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		h := sha256.Sum256(key.Marshal())
		fingerprint := base64.StdEncoding.EncodeToString(h[:])
		fmt.Printf(">> Security: Server Host Key Fingerprint (SHA-256): %s\n", fingerprint)
		return nil
	}

	config := &ssh.ClientConfig{
		User:            s.User,
		Auth:            authMethods,
		HostKeyCallback: hostKeyCallback,
		Timeout:         10 * time.Second,
	}

	client, err := ssh.Dial("tcp", address, config)
	if err != nil {
		fmt.Printf(">> Network: SSH Handshake Failed: %v\n", err)
		return core.ErrAuthFailed
	}

	s.SshClient = client
	fmt.Println(">> Network: Authenticated & Channel Encrypted.")

	return nil
}

// OpenSFTP initializes the SFTP subsystem on top of the SSH tunnel.
// This is distinct from Connect() because sometimes we just want Shell, not files.
func (s *SftpSession) OpenSFTP() error {
	if s.SshClient == nil {
		return core.ErrConnectionFailed
	}

	fmt.Println(">> Network: Requesting SFTP subsystem...")

	// Create the SFTP client using the existing SSH connection
	client, err := sftp.NewClient(s.SshClient)
	if err != nil {
		fmt.Printf(">> Network: Failed to open SFTP subsystem: %v\n", err)
		return core.ErrConnectionFailed
	}

	s.SftpClient = client
	fmt.Println(">> Network: SFTP Subsystem Active. Ready for I/O.")
	return nil
}

// Close disconnects everything politely.
func (s *SftpSession) Close() {
	if s.SftpClient != nil {
		s.SftpClient.Close()
	}
	if s.SshClient != nil {
		s.SshClient.Close()
	}
}
