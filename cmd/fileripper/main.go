// This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
// If a copy of the MPL was not distributed with this file, You can obtain one at
// https://mozilla.org/MPL/2.0/.

package main

import (
	"fmt"
	"os"
	"strconv"

	"fileripper/internal/core"
	"fileripper/internal/network"
	"fileripper/internal/pfte"
)

// main is the entry point.
// Go handles memory, so we focus on logic.
func main() {
	fmt.Println("FileRipper v0.0.1 - Powered by PFTE")

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
		// Usage: fileripper transfer <host> <port>
		if len(os.Args) < 4 {
			fmt.Println("Error: Missing host or port.")
			fmt.Println("Usage: fileripper transfer <host> <port>")
			return
		}

		host := os.Args[2]
		portStr := os.Args[3]
		port, err := strconv.Atoi(portStr)
		if err != nil {
			fmt.Println("Error: Invalid port number.")
			return
		}

		fmt.Printf(">> CLI Transfer mode engaged. Target: %s:%d\n", host, port)

		// 1. Init Network
		session := network.NewSession(host, port)
		defer session.Close()

		// 2. Test Connection (Handshake)
		if err := session.Connect(); err != nil {
			os.Exit(1)
		}

		// 3. Start Engine
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
  transfer       <host> <port> -> Test connection and engine
`)
}