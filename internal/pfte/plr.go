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
		assignedSession := sessions[i%sessionCount]

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
