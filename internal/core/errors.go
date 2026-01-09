// This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
// If a copy of the MPL was not distributed with this file, You can obtain one at
// https://mozilla.org/MPL/2.0/.

package core

import "errors"

// Common errors for the application.
// We define them here to avoid magic strings in the UI.
var (
	ErrConnectionFailed = errors.New("connection_failed")
	ErrHostUnreachable  = errors.New("host_unreachable")
	ErrAuthFailed       = errors.New("authentication_failed")
	
	// PFTE specific
	ErrPipelineStalled  = errors.New("pipeline_stalled")
	
	// System
	ErrUnknownCommand   = errors.New("unknown_command")
)