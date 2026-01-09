// This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
// If a copy of the MPL was not distributed with this file, You can obtain one at
// https://mozilla.org/MPL/2.0/.

package pfte

import (
	"fmt"

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
	Queue []string 
}

func NewEngine() *Engine {
	return &Engine{
		Mode:  ModeBoost, 
		Queue: make([]string, 0),
	}
}

func (e *Engine) StartTransfer(session *network.SftpSession) error {
	batchSize := BatchSizeConservative
	modeName := "Conservative"

	if e.Mode == ModeBoost {
		batchSize = BatchSizeBoost
		modeName = "Boost"
	}

	fmt.Printf(">> PFTE Engine started in %s mode.\n", modeName)
	fmt.Printf(">> Strategy: Launching %d concurrent workers (Goroutines).\n", batchSize)

	return nil
}