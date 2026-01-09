// This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
// If a copy of the MPL was not distributed with this file, You can obtain one at
// https://mozilla.org/MPL/2.0/.

package main

import (
	"fmt"
	"os"
	"strconv"

	// Local imports
	"fileripper/internal/core"
	"fileripper/internal/network"
	"fileripper/internal/pfte"
)

// main is the entry point.
func main() {
	fmt.Println("FileRipper v0.0.1 - Powered by PFTE (Go Edition)")

	if len(os.Args) < 2 {
		printUsage()
		return
	}

	command := os.Args[1]

	switch command {
	case "start-server":
		fmt.Println(">> Starting PFTE Server loop (Daemon mode)...")
		// TODO: Init API server here

	case "transfer":
		// Usage: fileripper transfer <host> <port> <user> <password>
		// Note: Passing password via CLI args is bad practice for prod (history logs),
		// but acceptable for this alpha test phase.
		if len(os.Args) < 6 {
			fmt.Println("Error: Missing arguments.")
			fmt.Println("Usage: fileripper transfer <host> <port> <user> <password>")
			return
		}

		host := os.Args[2]
		portStr := os.Args[3]
		user := os.Args[4]
		password := os.Args[5]

		port, err := strconv.Atoi(portStr)
		if err != nil {
			fmt.Println("Error: Invalid port number.")
			return
		}

		fmt.Printf(">> CLI Transfer mode engaged. Target: %s@%s:%d\n", user, host, port)

		// 1. Init Network with Auth credentials
		session := network.NewSession(host, port, user, password)
		defer session.Close()

		// 2. Connect & Authenticate (SHA-256 check happens here)
		if err := session.Connect(); err != nil {
			os.Exit(1)
		}

		// 3. Start Engine
		// We pass the authenticated session to the engine now.
		engine := pfte.NewEngine()
		if err := engine.StartTransfer(session); err != nil {
			fmt.Printf("Error during transfer: %v\n", core.ErrPipelineStalled)
		}
		
	default:
		fmt.Printf("Error: %v: %s\n", core.ErrUnknownCommand, command)
		printUsage()
	}
}

func printUsage() {
	fmt.Println(`
Usage: fileripper [command] [args]

Commands:
  start-server   Daemon mode (API for Flutter UI)
  transfer       <host> <port> <user> <password>
`)
}