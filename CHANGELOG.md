# Changelog

All notable changes to this project will be documented in this file.

## [v0.0.1] - 2026-05-11

### Initial Release

This is the first official release of gos7, a pure Go implementation of the Siemens S7 protocol.

### Features

#### AG (Programmer/PLC Communication)
- Read/Write Data Block (DB)
- Read/Write Merkers (MB)
- Read/Write Inputs (EB)
- Read/Write Outputs (AB)
- Read/Write Timer (TM)
- Read/Write Counter (CT)
- Multiple Read/Write Area operations
- Batch Read/Write Areas (ReadAreas/WriteAreas)
- Get Block Info

#### PG (PLC Control)
- Hot start/Cold start/Stop PLC
- Get CPU PLC status
- List available blocks in PLC
- Set/Clear password for session
- Get CPU protection and CPU Order code
- Get CPU/CP Information
- Read/Write clock for the PLC

#### Helpers
- Get/set values for byte arrays supporting: bit, int, word, dword, uint, real, time, counter

### Supported Communication
- TCP (stable)
- Serial (PPI, MPI) (under construction)

### Requirements
- Go 1.13 or higher (recommended: Go 1.21)