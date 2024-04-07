# Status Of Bumblebee Project
_Last updated Nov 3, 2023_

Currently working on Phase 1.  See ROAD_MAP.md for details on phase descriptions.

## Phase 1 is now done.  See [the Road Map](ROAD_MAP.md) for details on phase descriptions.
Relating to Phase 1...
- All necessary CLI commands are completed.  
- All use cases for encrypt/decrypt paths are completed and tested.
- All input and output code paths are supported.
- All input and output encodings are working.
- "bumblebee read" functionality provided for pipe support and/or replacement on Windows
- Tested on Mac, Windows and Linux.

## Possible remaining CLI functionality. Need to determine priorities and value 
- Error outputs need to be cleaned up.  Currently, most errors are wrapped on returns and do not
  print out very well.  They should be easier to read.
- Add more debug output
- Consider key rotation of some type, an ability to change/update key sets and possibly re-key bundles.  
This would also need to be considered for future server/service efforts.  Perhaps, instead, the user would just 
deprecate the current keyset and create a new identity?
- Add more unit tests for non-security critical paths 
- Add a CLI command "verify"?  This would simply verify the sending user identity for a bundle without having to
open and extract the bundle.
- Add a CLI command that would cycle through a set of users and attempt to determine the sending user
for a particular bundle. Maybe a command like "who"?
- Add more specific unit tests for failure and error flows
- Add more unit tests to validate all encodings that have only been manually validated
- Finish distribution analysis utility to confirm that output emissions do not favor specific binary 
patterns or byte values relating to location biases 
- Installers for each supported platform instead of just using Brew? 
- Review code to find any scenarios that might need missing "wipe" functionality. Should be covered, but maybe double/triple check
- Error outputs need to be cleaned up.  Currently, most errors are wrapped on returns and do not
print out very well.  They should be easier to read.

## Known issues
Issues with "make" utility on Windows. Possible workarounds...
- Run builds on Windows from WSL -- **Not tested**
- Run builds on Windows using "go build" commands, but no metadata injected into version structure
- Maybe provide a batch file or script of some type for building without MAKEFILE
