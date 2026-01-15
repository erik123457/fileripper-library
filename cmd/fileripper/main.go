// This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
// If a copy of the MPL was not distributed with this file, You can obtain one at
// https://mozilla.org/MPL/2.0/.

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"fileripper/internal/core"
	"fileripper/internal/network"
	"fileripper/internal/pfte"
	"fileripper/internal/server"
)

const SessionCount = 2 // Sessions (adjust if necessary)

func main() {
	fmt.Println("FileRipper v0.4.0 - Powered by PFTE ")

	if len(os.Args) < 2 {
		printUsage()
		return
	}

	command := os.Args[1]

	switch command {
	case "start-server":
		port := 2935 // WTF was I thinking with 112?
		if len(os.Args) > 2 {
			p, err := strconv.Atoi(os.Args[2])
			if err == nil {
				port = p
			}
		}
		server.StartDaemon(port)

	case "transfer":
		handleTransferCLI(os.Args)
		
	default:
		fmt.Printf("Error: %v: %s\n", core.ErrUnknownCommand, command)
		printUsage()
	}
}

func handleTransferCLI(args []string) {
	if len(args) < 6 {
		fmt.Println("Error: Missing arguments.")
		fmt.Println("Usage: fileripper transfer <host> <port> <user> <pass> [--upload <local> <remote_dest> | --download <remote>]")
		return
	}

	host := args[2]
	portStr := args[3]
	user := args[4]
	password := args[5]
	
	operation := "DOWNLOAD" 
	sourcePath := "."       
	destPath := "." // Default destination (Root)

	if len(args) > 6 {
		mode := strings.ToLower(args[6])
		if mode == "--upload" {
			operation = "UPLOAD"
			if len(args) > 7 {
				sourcePath = args[7]
			}
			// Capture remote destination if provided
			if len(args) > 8 {
				destPath = args[8]
			}
		} else if mode == "--download" {
			operation = "DOWNLOAD"
			if len(args) > 7 {
				rawPath := args[7]
				sourcePath = filepath.ToSlash(rawPath)
			}
		}
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		fmt.Println("Error: Invalid port number.")
		return
	}

	targetDisplay := sourcePath
	if operation == "UPLOAD" {
		targetDisplay = fmt.Sprintf("%s -> %s", sourcePath, destPath)
	}
	
	fmt.Printf(">> CLI Mode: %s. Target: %s (%s@%s:%d)\n", operation, targetDisplay, user, host, port)

	// --- DUAL SESSION INIT ---
	var sessions []*network.SftpSession
	fmt.Printf(">> Network: Establishing %d parallel tunnels...\n", SessionCount)

	for i := 0; i < SessionCount; i++ {
		sess := network.NewSession(host, port, user, password)
		if err := sess.Connect(); err != nil {
			fmt.Printf("Error connecting session #%d: %v\n", i+1, err)
			os.Exit(1)
		}
		if err := sess.OpenSFTP(); err != nil {
			fmt.Printf("Error opening SFTP #%d: %v\n", i+1, err)
			os.Exit(1)
		}
		sessions = append(sessions, sess)
	}
	
	defer func() {
		for _, s := range sessions { s.Close() }
	}()

	engine := pfte.NewEngine()

	// --- CLI DASHBOARD ---
	stopMonitor := make(chan bool)
	go func() {
		ticker := time.NewTicker(200 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-stopMonitor:
				fmt.Println("\n>> Monitor stopped.")
				return
			case <-ticker.C:
				stats := pfte.GlobalMonitor.GetStats()
				if stats.IsRunning {
					fmt.Printf("\r[PROGRESS] %.1f%% | Speed: %.2f MB/s | Files: %d/%d | Current: %s          ", 
						stats.ProgressPercent, stats.SpeedMBs, stats.FilesDone, stats.TotalFiles, limitString(stats.CurrentFile, 20))
				}
			}
		}
	}()

	// Pass both Source and Dest
	if err := engine.StartTransfer(sessions, operation, sourcePath, destPath); err != nil {
		fmt.Printf("\nError during transfer: %v\n", err)
	}
	
	stopMonitor <- true
	time.Sleep(100 * time.Millisecond) 
}

func limitString(s string, max int) string {
	if len(s) > max {
		return s[:max] + "..."
	}
	return s
}

func printUsage() {
	fmt.Println(`
Usage: fileripper [command] [args]

Commands:
  start-server [port]   Start REST API Daemon
  transfer              <host> <port> <user> <pass> [--upload <local> <remote_dest> | --download <remote>]
`)
}