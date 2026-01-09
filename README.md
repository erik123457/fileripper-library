# FileRipper (Experimental)

![License](https://img.shields.io/badge/License-MPL%202.0-red.svg) ![Zig](https://img.shields.io/badge/Made%20with-Zig-orange.svg) ![Status](https://img.shields.io/badge/Status-Experimental-yellow.svg)

Standard SFTP clients are slow because they place too much emphasis on sequential processing. FileRipper is designed to saturate the pipeline.

It uses the PFTE engine to manage parallel pipelines, avoiding the slowdowns caused by typical sequential bottlenecks.

## Main Features (PFTE)

- **Adaptive Request Grouping:** Instead of processing requests one by one, we send batches of up to 64 requests.

- **No Overhead:** Written in Zig. No garbage collection or hidden allocations in the active path.

- **Split Architecture:** The kernel is independent. The user interface is for ease of use.

##Quick Start

1. Install the Zig compiler (0.11.x+). 2. Cloning and Compiling:

$ git clone https://github.com/axelcaruso/fileripper-library.git
$ cd fileripper
$ zig ​​build -Doptimize=ReleaseFast

## Notes for Contributors

- Run 'zig fmt' or your request will be rejected.

- Keep speed in mind. If it's not O(1) or O(n), think twice.

- If the server is old or unstable, use Conservative mode to avoid triggering the IPS.

--- EXPERIMENTAL: Use at your own risk.