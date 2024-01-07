# Status Of Bumblebee Project
_Last updated Nov 3, 2023_

Currently working on Phase 1.  See ROAD_MAP.md for details on phase descriptions.

## Current General Status
- Bumblebee is MVP functional.  All necessary CLI commands are complete.  All use cases for encrypt/decrypt 
paths are complete.
- All input and output code paths are supported.
- All input and output encodings are working.
- "bumblebee read" functionality provided for pipe support and/or replacement on Windows
- Tested on Mac, Windows and Linux.

## What's left to do
- Add more debug output
- Consider key rotation of some type, an ability to change/update key sets and possibly re-key bundles.  
This would also need to be considered for future server/service efforts.  Perhaps, instead, just 
deprecate current keyset and create a new one.
- Add more unit tests for not critical paths 
- Support verify command
- Support backup and restore commands
- Add unit tests for failure and error flows
- Add unit tests to validate all encodings, which have been manually validated
- Finish distribution analysis utility to confirm that output emissions do not favor specific binary 
patterns or byte values relating to location biases 
- Installers for each supported platform
- Scan code for missing wipes where sensitive values are handled... should be covered, but maybe double/triple check
- Currently, Bee only supports a single file or text input.  Probably, it should support encrypting whole directories.
Or maybe just multiple secrets per bundle in some form. Due to certain complexities, 
that may be in a follow-on version.
- Error outputs need to be cleaned up.  Currently, most errors are wrapped on returns and do not
print out very well.  They should be easier to read.

## Known issues
Issues with "make" utility on Windows. Possible workarounds...
- Run builds on Windows from WSL -- **Not tested**
- Run builds on Windows using "go build" commands, but no metadata injected into version structure
- Maybe provide a batch file or script of some type for building without MAKEFILE
