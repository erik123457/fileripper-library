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

package core

import "errors"

// Common errors for the application.
// We define them here to avoid magic strings in the UI.
// PS: It will be improved; for now, works :)
var (
	ErrConnectionFailed = errors.New("connection_failed")
	ErrHostUnreachable  = errors.New("host_unreachable")
	ErrAuthFailed       = errors.New("authentication_failed")

	// PFTE specific
	ErrPipelineStalled = errors.New("pipeline_stalled")

	// System
	ErrUnknownCommand = errors.New("unknown_command")
)
