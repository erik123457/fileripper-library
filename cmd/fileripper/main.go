// This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
// If a copy of the MPL was not distributed with this file, You can obtain one at
// https://mozilla.org/MPL/2.0/.

package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"fileripper/internal/core"
	"fileripper/internal/network"
	"fileripper/internal/pfte"
	"fileripper/internal/server"
)

func main() {
	fmt.Println("FileRipper v0.3.0 - Powered by PFTE Engine")

	if len(os.Args) < 2 {
		printUsage()
		return
	}

	command := os.Args[1]

	switch command {
	case "start-server":
		port := 112
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
		fmt.Println("Usage: fileripper transfer <host> <port> <user> <password> [--upload <folder> | --download]")
		return
	}

	host := args[2]
	portStr := args[3]
	user := args[4]
	password := args[5]
	
	operation := "DOWNLOAD" // Default
	targetPath := ""

	if len(args) > 7 && strings.ToLower(args[6]) == "--upload" {
		operation = "UPLOAD"
		targetPath = args[7] // The folder to upload
	} else if len(args) > 6 && strings.ToLower(args[6]) == "--download" {
		operation = "DOWNLOAD"
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		fmt.Println("Error: Invalid port number.")
		return
	}

	fmt.Printf(">> CLI Mode: %s. Target: %s@%s:%d\n", operation, user, host, port)

	session := network.NewSession(host, port, user, password)
	defer session.Close()

	if err := session.Connect(); err != nil {
		os.Exit(1)
	}

	if err := session.OpenSFTP(); err != nil {
		os.Exit(1)
	}

	engine := pfte.NewEngine()

	// --- CLI DASHBOARD (Background Monitor) ---
	// This goroutine prints the live stats to the console while the engine runs.
	stopMonitor := make(chan bool)
	go func() {
		ticker := time.NewTicker(200 * time.Millisecond) // Update 5 times a second
		defer ticker.Stop()
		
		for {
			select {
			case <-stopMonitor:
				fmt.Println("\n>> Monitor stopped.")
				return
			case <-ticker.C:
				stats := pfte.GlobalMonitor.GetStats()
				if stats.IsRunning {
					// \r moves cursor to start of line, allowing overwrite
					fmt.Printf("\r[PROGRESS] %.1f%% | Speed: %.2f MB/s | Files: %d/%d | Current: %s          ", 
						stats.ProgressPercent, 
						stats.SpeedMBs, 
						stats.FilesDone, 
						stats.TotalFiles,
						limitString(stats.CurrentFile, 20),
					)
				}
			}
		}
	}()

	// Start the Engine (Blocking call)
	if err := engine.StartTransfer(session, operation, targetPath); err != nil {
		fmt.Printf("\nError during transfer: %v\n", core.ErrPipelineStalled)
	}
	
	// Stop the monitor cleanly
	stopMonitor <- true
	time.Sleep(100 * time.Millisecond) // Let the newline print
}

// Helper to prevent the UI from breaking if filename is too long
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
  transfer              <host> <port> <user> <pass> [--upload <local_folder> | --download]
`)
}