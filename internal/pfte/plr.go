// This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
// If a copy of the MPL was not distributed with this file, You can obtain one at
// https://mozilla.org/MPL/2.0/.

package pfte

import (
	"fmt"
	"sync"
	"time"

	"fileripper/internal/network"
)

// WorkerPool manages the swarm of goroutines.
type WorkerPool struct {
	Concurrency int
	Queue       *JobQueue
	Wg          sync.WaitGroup
}

func NewWorkerPool(concurrency int, queue *JobQueue) *WorkerPool {
	return &WorkerPool{
		Concurrency: concurrency,
		Queue:       queue,
	}
}

// StartUnleash fires up the goroutines using ROUND ROBIN session balancing.
func (wp *WorkerPool) StartUnleash(sessions []*network.SftpSession) {
	sessionCount := len(sessions)
	fmt.Printf(">> PLR: Unleashing %d workers across %d tunnels...\n", wp.Concurrency, sessionCount)
	
	GlobalMonitor.SetRunning(true)
	start := time.Now()

	for i := 0; i < wp.Concurrency; i++ {
		wp.Wg.Add(1)
		
		// Load Balance: Worker 0 -> Sess 0, Worker 1 -> Sess 1, Worker 2 -> Sess 0...
		assignedSession := sessions[i % sessionCount]

		go func(workerID int, sess *network.SftpSession) {
			defer wp.Wg.Done()
			
			for {
				job := wp.Queue.Pop()
				if job == nil {
					return 
				}

				GlobalMonitor.SetCurrentFile(job.RemotePath)

				var err error
				if job.Operation == "DOWNLOAD" {
					err = DownloadFileWithProgress(sess, job.RemotePath, job.LocalPath)
				} else if job.Operation == "UPLOAD" {
					err = UploadFileWithProgress(sess, job.LocalPath, job.RemotePath)
				}

				if err != nil {
					// Concise logging to avoid console spam
					// log.Printf("[Worker %d] Fail: %v", workerID, err)
					// Simple retry logic is inside transfer functions
					continue
				}

				GlobalMonitor.IncFileDone()
			}
		}(i, assignedSession)
	}

	wp.Wg.Wait()
	GlobalMonitor.SetRunning(false)
	
	duration := time.Since(start)
	fmt.Printf(">> PLR: Batch complete in %v.\n", duration)
}