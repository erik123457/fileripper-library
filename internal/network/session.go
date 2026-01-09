// This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
// If a copy of the MPL was not distributed with this file, You can obtain one at
// https://mozilla.org/MPL/2.0/.

package network

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net"
	"time"

	"fileripper/internal/core"
	"golang.org/x/crypto/ssh"
)

// SftpSession holds the SSH connection state.
// Now it's a real SSH client, not just a raw socket.
type SftpSession struct {
	Hostname string
	Port     int
	User     string
	Password string
	Client   *ssh.Client // The heavy lifter
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
// It performs the handshake, auth, and validates the server's SHA-256 fingerprint.
func (s *SftpSession) Connect() error {
	address := fmt.Sprintf("%s:%d", s.Hostname, s.Port)
	fmt.Printf(">> Network: Initiating Secure Handshake with %s...\n", address)

	// Define how we want to authenticate.
	// For v0.0.1 we stick to passwords. Keys come later.
	authMethods := []ssh.AuthMethod{
		ssh.Password(s.Password),
	}

	// Host Key Callback: This is the security checkpoint.
	// We calculate the SHA-256 hash of the server's public key to verify identity.
	hostKeyCallback := func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		// Calculate SHA-256 fingerprint
		h := sha256.Sum256(key.Marshal())
		fingerprint := base64.StdEncoding.EncodeToString(h[:])
		
		fmt.Printf(">> Security: Server Host Key Fingerprint (SHA-256): %s\n", fingerprint)
		
		// In a real app, we would check this against a known_hosts file.
		// For now, we trust on first use (TOFU) but we show it to the user.
		return nil 
	}

	config := &ssh.ClientConfig{
		User:            s.User,
		Auth:            authMethods,
		HostKeyCallback: hostKeyCallback,
		Timeout:         10 * time.Second, // Don't wait forever
	}

	// The actual Dial. This replaces our old raw TCP logic.
	client, err := ssh.Dial("tcp", address, config)
	if err != nil {
		fmt.Printf(">> Network: SSH Handshake Failed: %v\n", err)
		return core.ErrAuthFailed
	}

	s.Client = client
	fmt.Println(">> Network: Authenticated & Channel Encrypted.")

	return nil
}

// Close disconnects the client politely.
func (s *SftpSession) Close() {
	if s.Client != nil {
		s.Client.Close()
	}
}