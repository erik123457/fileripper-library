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
	"hash/crc32"
	"io"
	"os"
)

// CalculateChecksum computes the CRC32 hash of a file.
// We use CRC32 because SHA256 is too slow for high-throughput transfer checks.
// We just want to know if the file got corrupted, not sign a contract.
func CalculateChecksum(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// IEEE is the standard used by Ethernet, Zip, etc. Fast and reliable.
	hasher := crc32.NewIEEE()

	// Copy the file content into the hasher in chunks (32KB buffer usually)
	// efficiently without loading the whole file into RAM.
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}

	// Return as hex string for easy comparison
	return fmt.Sprintf("%x", hasher.Sum32()), nil
}
