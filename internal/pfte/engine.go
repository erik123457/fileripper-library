// This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
// If a copy of the MPL was not distributed with this file, You can obtain one at
// https://mozilla.org/MPL/2.0/.

package pfte

import (
	"fmt"
	"os"
	"path/filepath"
	"fileripper/internal/network"
)

const (
	BatchSizeBoost        = 64
	BatchSizeConservative = 2
)

type TransferMode int

const (
	ModeBoost        TransferMode = iota 
	ModeConservative                     
)

type Engine struct {
	Mode  TransferMode
	Queue *JobQueue 
}

func NewEngine() *Engine {
	return &Engine{
		Mode:  ModeBoost, 
		Queue: NewQueue(),
	}
}

// StartTransfer is now bidirectional.
// mode: "DOWNLOAD" (default) or "UPLOAD"
// targetPath: The local folder to upload (if mode is UPLOAD)
func (e *Engine) StartTransfer(session *network.SftpSession, operation string, targetPath string) error {
	if session.SftpClient == nil {
		return fmt.Errorf("sftp_client_not_initialized")
	}

	concurrency := BatchSizeConservative
	if e.Mode == ModeBoost {
		concurrency = BatchSizeBoost
	}

	queuedCount := int64(0)
	totalBytes := int64(0)

	// --- UPLOAD LOGIC ---
	if operation == "UPLOAD" {
		fmt.Printf(">> PFTE: Scanning local directory '%s' for mass upload...\n", targetPath)
		
		err := filepath.Walk(targetPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				// We create remote directories on the fly or let sftp handle it?
				// For v0.2, let's assume flat structure or pre-existing dirs.
				// Better: Just skip dir entries, file creation handles structure if we implement mkdir (later).
				return nil 
			}

			// Calculate remote path relative to the target folder
			// e.g. local: main/src/code.go -> remote: ./main/src/code.go
			// For simplicity in this version, we upload to remote root retaining the filename.
			// Or better: upload to a folder named as the source.
			
			// Simple Logic: Local "main/file.txt" -> Remote "./file.txt" (Flattening for now for safety)
			// TODO: Implement recursive directory creation on remote.
			remotePath := info.Name() 

			e.Queue.Add(&TransferJob{
				LocalPath:  path,
				RemotePath: remotePath,
				Operation:  "UPLOAD",
			})
			
			queuedCount++
			totalBytes += info.Size()
			return nil
		})

		if err != nil {
			return err
		}

	// --- DOWNLOAD LOGIC ---
	} else {
		// Default to scanning remote root
		localDir := "dump"
		if _, err := os.Stat(localDir); os.IsNotExist(err) {
			os.Mkdir(localDir, 0755)
		}
		
		fmt.Println(">> PFTE: Scanning remote root for download...")
		files, err := session.SftpClient.ReadDir(".")
		if err != nil {
			return err
		}

		for _, file := range files {
			if file.IsDir() { continue }
			
			localPath := filepath.Join(localDir, file.Name())
			e.Queue.Add(&TransferJob{
				LocalPath:  localPath,
				RemotePath: file.Name(),
				Operation:  "DOWNLOAD",
			})
			queuedCount++
			totalBytes += file.Size()
		}
	}

	fmt.Printf(">> PFTE: Job ready. Files: %d, Total Size: %d bytes.\n", queuedCount, totalBytes)
	
	GlobalMonitor.Reset(queuedCount, totalBytes)

	if queuedCount == 0 {
		return nil
	}

	// Launch the swarm
	workerPool := NewWorkerPool(concurrency, e.Queue)
	workerPool.StartUnleash(session)

	return nil
}