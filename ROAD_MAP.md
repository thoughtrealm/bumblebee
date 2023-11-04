# Road Map Of Bumblebee Project
Last updated Nov 2, 2023

This is a rough idea of possible future work priorities.  It is possible that
priorities may shift, which would mean the ordering of these phases could change.

## Phase 1 - CLI (Command Line Interpreter)
_Phase 1 is approximately 85% complete.  It is 100% MVP complete._

**The primary goal of Phase 1 is to deliver a CLI application that provides at least 
the following functionality:**
- Initialization of a local environment that contains random cryptographic components
for the encrypting and decrypting of data.

- Provide an encrypting/decrypting system that has the following properties:
  - Utilizes both symmetric and asymmetric encryption
  - The user should not have to know or provide the symmetric key components, something like
  an offline TLS.
  - The user should only have to exchange public keys with other users in order to securely share secrets.
  - Private keys should never be transmitted and should be stored only in a protected 
  space in their local environment.
  - It should support encrypting of large streams, such as gigabyte length files.

- Allow the user to securely and easily maintain a store of public keys for exchanging
data with other users.
- The user should have the option of providing a symmetric key for encrypting the storage of
key sets.  This key would have to be input every time the app runs, or it would have to be provided
via some alternate input like an environment variable, etc.
- Provide simple functionality that makes it easy to do basic functions, like adding users,
  removing users, editing users, etc.
- Support sharing files, as well as simple manual text entry
- Support encrypting outputs in text safe formats that can be copied off the screen and
safely shared via pasting into open environments.  These could be shared in messaging apps or 
an email body, generally any medium that allows text to be shared between users.
- Support multiple profiles and multiple local key sets to make it easier to exchange data 
with different groups or to use different keys for different purposes.
- Support encrypting and decrypting files for personal storage or pre-encrypting prior to posting
files on some storage medium, like iCloud, OneDrive, etc.
- Provide easy functionality to update the user database with new user keys as needed,
if someone should need to change their keys
- It should be easy to back up and restore the local environment, including key storage.


## Phase 2 - Provide A Key Management Server
The primary goal of Phase 2 is to provide a key management server that provides at least the 
following functionality:
- Securely stores public keys in some form.  Exact DB or details are TBD, but the local keystore
mechanisms might be sufficient.  They could support fairly large user sets, but would
incur relatively high latency on rights with large user sets.  
- Allow optional support for TLS
- Use a cipher exchange approach similar to the bee asymmetric+symmetric encoding format for
when no TLS is configured.  This should be fine for any scenario, including public transports,
but definitely should be fine for local or internal networks.
- It would be managed by a user admin role.
- User profiles could be loaded in bulk by the maintainer(s) using some file format, maybe
comma delimited, JSON, etc.
- Users would push their public keys to this service as a K/V store. The user ID could be their
email address or anything unique.
- When encrypting files locally with Bee, if the reference target user was not in the local
key store, then it would request the key from the server.
- Local user would "sync" the environment, which would mean the local system requested changes
to local key stores for any updates that occur on the server, like the user changed their key, etc.

### The Unique Issues In Supporting A "User Groups" Feature:
DISCUSS THIS USE CASE HERE AND THE POSSIBLE CONCERNS IT WOULD CREATE, POSSIBLE WAYS TO SOLVE THIS
BY SWAPPING STREAM HEADERS, ETC.

## Phase 3 - Service For Distribution Of Secrets
The goal of Phase 3 is to provide a service that can easily distribute secrets between
users.  The feature set is TBD, but this would surely be some kind of store and forward mechanism,
not a real time exchage.  The local user would configure a server for the profile key stores. This
would allow the user to "push" encrypted secrets to a distribution service.  The receiving 
user would sync or "pull" the secret down and decrypt it.

This would likely be functionality added to the key management server, so you don't have to
support two servers.

This MIGHT need to allow multiple server instances for load balancing and other needs.  If so,
the store paths in the server environment would need to support concurrent access and shared
by all instances.

The data storage mechanism on the server is TBD.  It could be a file system for the persistence
that also incorporates a DB for fast metadata access.

## Phase 4 - Provide a User Interface
The goal of Phase 4 is to provide a UI in some form, possibly through a local web service
or via a native library like Fyne.  This is TBD.  But a major goal here, should a browser UI
be utilized, would be security.  Maybe, the CLI spawns a local web UI and communicates over a
web socket.  That approach should be considered along with native libs.

## Phase 4 - Mobile Support
Phase 4 will provide the Bee feature set on mobile platforms.  The exact solution and
requirements are TBD.