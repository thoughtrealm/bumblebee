# Status Of Bumblebee Project
_Last updated Nov 3, 2023_

Currently working on Phase 1.  See ROAD_MAP.md for details on phase descriptions.

## Current General Status
- Bee is MVP functional.  All necessary CLI commands are complete.  All use cases for encrypt/decrypt 
paths are complete.
- All input and output code paths are supported.
- All input and output encodings are working.
- "bee read" functionality provided for pipe support and/or replacement on Windows
- Tested on Mac, Windows and Linux.

## What's left to do
- Add more debug output
- Add more unit tests for not critical paths 
- Add unit tests for failure and error flows
- Add unit tests to validate all encodings, which were validated manually for now
- Finish distribution analysis utility to confirm that output emissions do not favor specific binary 
patterns or byte location biases 
- Installers for each supported platform

## Known issues
Issues with "make" utility on Windows. Possible workarounds...
- Run builds on Windows from WSL -- **Not tested**
- Run builds on Windows using "go build" commands, but no metadata inserted into version structure
- Maybe provide a batch file or script of some type for building without MAKEFILE
