// This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
// If a copy of the MPL was not distributed with this file, You can obtain one at
// https://mozilla.org/MPL/2.0/.

package pfte

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"sync/atomic"

	"fileripper/internal/network"
)

const (
	BatchSizeBoost        = 64
	BatchSizeConservative = 2
	DirCreationWorkers    = 8 // Optimized for Phase 1
)

type TransferMode int

const (
	ModeBoost TransferMode = iota
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

// StartTransfer enforces explicit directory creation (Phase 1) then file transfer (Phase 2).
func (e *Engine) StartTransfer(session *network.SftpSession, operation string, targetPath string) error {
	if session.SftpClient == nil {
		return fmt.Errorf("sftp_client_not_initialized")
	}

	concurrency := BatchSizeConservative
	if e.Mode == ModeBoost {
		concurrency = BatchSizeBoost
	}

	// --- UPLOAD LOGIC ---
	if operation == "UPLOAD" {
		// Convert input to Absolute Path immediately.
		absTarget, err := filepath.Abs(targetPath)
		if err != nil {
			return fmt.Errorf("failed to get absolute path: %v", err)
		}

		fmt.Printf(">> PFTE: Analyzing local source '%s'...\n", absTarget)

		baseDir := filepath.Dir(absTarget)

		var foldersToCreate []string
		var filesToTransfer []*TransferJob
		
		totalBytes := int64(0)

		// STEP 1: SCAN
		err = filepath.Walk(absTarget, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				fmt.Printf(">> Warning: Error accessing %s: %v\n", path, err)
				return nil
			}

			relPath, err := filepath.Rel(baseDir, path)
			if err != nil {
				return err
			}
			
			remotePath := filepath.ToSlash(relPath)

			if info.IsDir() {
				if remotePath != "." && remotePath != "" {
					foldersToCreate = append(foldersToCreate, remotePath)
				}
			} else {
				filesToTransfer = append(filesToTransfer, &TransferJob{
					LocalPath:  path,
					RemotePath: remotePath,
					Operation:  "UPLOAD",
				})
				totalBytes += info.Size()
			}
			return nil
		})

		if err != nil {
			return err
		}

		// STEP 2: PHASE 1 - CREATE STRUCTURE (PARALLEL x8)
		sort.Slice(foldersToCreate, func(i, j int) bool {
			return len(foldersToCreate[i]) < len(foldersToCreate[j])
		})

		dirCount := len(foldersToCreate)
		if dirCount > 0 {
			fmt.Printf(">> PFTE: Phase 1 - Creating %d directories...\n", dirCount)
			
			dirChan := make(chan string, dirCount)
			var wg sync.WaitGroup
			var doneCount int32
			var printMu sync.Mutex 

			for _, d := range foldersToCreate {
				dirChan <- d
			}
			close(dirChan)

			for i := 0; i < DirCreationWorkers; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for dir := range dirChan {
						_ = session.SftpClient.MkdirAll(dir)
						current := atomic.AddInt32(&doneCount, 1)
						
						printMu.Lock()
						progress := float64(current) / float64(dirCount) * 100
						fmt.Printf("\r>> Structure: [%3.0f%%] Processed...                 ", progress)
						printMu.Unlock()
					}
				}()
			}
			wg.Wait()
			fmt.Println("\n>> PFTE: Structure verified.")
		}

		// STEP 3: PHASE 2 - QUEUE FILES
		fileCount := int64(len(filesToTransfer))
		if fileCount == 0 {
			fmt.Println(">> Warning: No files found to transfer.")
			return nil
		}

		fmt.Printf(">> PFTE: Phase 2 - Queuing %d files...\n", fileCount)
		for _, job := range filesToTransfer {
			e.Queue.Add(job)
		}

		GlobalMonitor.Reset(fileCount, totalBytes)

		workerPool := NewWorkerPool(concurrency, e.Queue)
		workerPool.StartUnleash(session)

		return nil

	// --- DOWNLOAD LOGIC ---
	} else {
		localDir := "dump"
		if _, err := os.Stat(localDir); os.IsNotExist(err) {
			os.Mkdir(localDir, 0755)
		}

		fmt.Println(">> PFTE: Scanning remote root for download...")
		files, err := session.SftpClient.ReadDir(".")
		if err != nil {
			return err
		}

		queuedCount := int64(0)
		totalBytes := int64(0)

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

		fmt.Printf(">> PFTE: Job ready. Files: %d, Total Size: %d bytes.\n", queuedCount, totalBytes)
		GlobalMonitor.Reset(queuedCount, totalBytes)
		
		if queuedCount > 0 {
			workerPool := NewWorkerPool(concurrency, e.Queue)
			workerPool.StartUnleash(session)
		}
		return nil
	}
}