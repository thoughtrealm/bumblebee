# A Threat Model Of The Bumblebee Project

## ***!!!NOTE: This document is still in process and is NOT complete at this time!!!***

## **I. Document Purpose**
Like all security-related projects, the _Bumblebee_ project consists of a number of trade-offs,
attempting to balance security objectives with practical software realities.  It is not perfect and 
you may find that it is more appropriate for some use-cases and usage scenarios than others.  

The purpose of this document is to provide you with an overview of the security concerns relating to
the _Bumblebee_ project.  This will include its components and technologies, strengths, recommended usage,
areas of concerns and risks, as well as how to address them when it is possible to do.  

The goal is for you or your agency to be sufficiently informed so that you are able to determine if Bumblebee
is an appropriate technology for your specific use case and environment.

## **II. Definitions**
The following are some terms we will define for the sake of document clarity.

### **A. Bumblebee**
_Bumblebee_ will refer to both the project itself, and any related binaries or artifacts that
provide Bumblebee's functionality. For example "_Bumblebee_" the project, "CLI" the binary,
and "_the Bumblebee CLI_" may be used interchangeably. When a distinction is required, it will be provided
using an explicit form, such as _"the CLI"_.

### **B. CLI or Command Line Interface**
A ***CLI***, or ***Command Line Interface*** is a binary that you can run using terminal commands.  It
may be invoked directly from a terminal interface, or it may be called from shell scripts or other
binaries.

### **C. Secret**
In the context of _Bumblebee_, a "_secret_" is simply something you do not want other people or systems
to know.  For example, your username and password for you online banking account is a ***secret***, in
that you definitely do not want other entities to know your banking credentials.  In that sense, any info
you want to keep private is a secret, wither it is credentials, documents, etc.

### **D. Bundles**
Bundles are data streams that store your secrets.  Generally these are files, but may be console outputs,
clipboard data, or pipe data.  Bundles refer specifically to the data sequence formats and constructions
that _Bumblebee_ uses to move secrets around.

## **III. A Description Of The Bumblebee Project**

### **A. Overview of the CLI**
Currently, _Bumblebee_ is implemented as a CLI (Command Line Interface).  Other implementations
may be available in the future and will be addressed in changes to this document as needed.

The primary objective of Bumblebee is to encrypt secrets, so that you can securely share them with other
_Bumblebee_ users, who are then able to decrypt them.  While _Bumblebee_ may have other capabilities in
addition to that functionality, sharing encrypted secrets is the primary use case provided by _Bumblebee_.

Currently, all _Bumblebee_ functionality is provided as a single binary in the form of a cross-platform
CLI.  The project conventions intentionally support the **ARM64** and **AMD64** targets
for the desktop environments of _Linux_, _Mac_ and _Windows_.  However, using the _Go_ compiler environment,
you may be able to create binaries for a number of other platforms.  This document will only consider the
intentionally supported targets and platforms.

The CLI is able to encrypt data from several input sources, as well as decrypt that data to
several output targets.

The supported input sources differ per platform, but are one of the following:
- Files
- Text entered directly from the console
- Clipboard input
- Piped input via Stdin

And the supported output targets may also differ per platform and include:
- Files
- Console Output
- Clipboard direct writes
- Piped output to Stdout (as a form of console output)

Given your particular environment, you may have additional options.  For example, _Linux_ may
provide commands and utilities that can re-direct inputs and outputs in various ways that _Windows_ or
_Mac_ may not provide, or may provide that functionality using different mechanisms.

### **B. Cryptographic Techniques and Algorithms Employed**
_Bumblebee_ utilizes a number of cryptographic techniques and algorithms to achieve its goals. How
these are used, as well as the source code origins used in _Bumblebee_, is described in detail
in the [Technical Details Document](TECHNICAL_DETAILS.md).  For this document, we will only mention the specific 
cryptographic components for the sake of describing threat concerns that may relate to them.

_Bumblebee_ utilizes cryptographic functionality in several areas.  Note that for all areas of 
cryptographic functionality, randomness is derived using the **Go** language crypto/rand package.

#### **1. Bundles Encrypted For Target Users**
_Bumblebee_ encrypts secrets into ***Bundles*** for a specific user.  _Bundles_ are encrypted using
***Hybrid Cryptography**.  _Bundles_ are composed of two sections; A header that is encrypted using
***asymmetric cryptography***, and a data payload that is encrypted using ***symmetric cryptography***.

For the _asymmetric functionality_, _Bumblebee_ makes use of **ed25519** for code signing and **curve25519**
for encryption.  The asymmetric functionality is taken from the **NATS** project's **NKEYS** package.
**NKEYS** also provides a utility wrapper around the **Go** language's crypto library implementation of
the **NaCl** box functionality. 

For the _symmetric functionality_, _Bumblebee_ makes use of **XChacha20-Poly1305** for encryption.  All data
payloads are encrypted with random keys, which are strengthened with **Argon2** and stored in the _asymmetric_
encrypted header.

#### **2. Data Encrypted For Local Use**
_Bumblebee_ provides functionality for encrypting sources for local use.  These encrypted local _Bundles_ are
created using the same techniques described above for user _Bundles_; however, the headers are encrypted using
a set of system read and write identities that are stored in the related local ***Profile***.  The
Bundles created in this process are not intended for distribution to other users, but would be encrypted
for backup storage to media or cloud systems. Other than the identity key pairs used, the cryptographic
approach is identical to user _Bundles_.

_Bundles_ from this process are generally intended for longer storage, though that may not always be the
case.  If you lose your profile in some way, such as from a system or drive failure, you will not be able
to decrypt these _bundles_.  Therefore, when using local keys in this way, it is recommended to harden the
profile space or securely back it up in some way.

#### **3. Export Files**
_Bumblebee_ provides functionality for exporting public keys.  These public keys are provided to other
users, which allows them to send you encrypted data. 

When exporting these public keys, you may optionally encrypt the public key data with a user-supplied
symmetric key.  There are no constraints or validations applied to this key, so you may any value you wish
for the initial value of the key.  The key is not provided with the export data in any form.  You are
responsible for providing the key value to anyone that you wish to have the export data.  If you lose or
forget the key, you and the targer user will not be able to decrypt the exported data.

When you elect to encrypt these files, _Bumblebee_ uses **XChacha20-Poly1305** to do so.  The key you provide
is strengthened with **Argon2** prior to encryption.

#### **4. Profile Identity Stores**
_Bumblebee_ creates an environment locally that is composed of some number of **profiles**.  These profiles
are stored in the local user config path that would be conventional for your operating system.

Each _profile_ stores identities and their related key data in two distinct identity stores. Each store is
a single file that is located in that profile's path.  These two files store all identities for that
profile, both your personal local identities and remote user identities.  

#### **4.a. Local Identities Store**
This store contains your local personal identities for that profile.  These contain the private keys that are
necessary to decrypt _bundles_ sent to you, as well as sign _bundles_ that you send to other users.

You have the option of encrypting this store with a user-supplied symmetric key. If you do so,
this store file will be encrypted as a single stream using **XChacha20-Poly1305**.  Your key will be
strengthened using **Argon2**.

When you choose to encrypt this file in this way, you will have to provide the key every time you run a
_Bumblebee_ command.  Currently, there are two options for this.  One is that you are prompted for it
when _Bumblbebee_ loads.  The other is that you may set an environment variable with the key and 
_Bumblebee_ will obtain the key from the environment variable.

If you forget this key or lose it, such as due to a drive or environment failure, you will not be
able to access your local identities anymore.  You are advised to securely back up this key in some way.

#### **4.b. User Identities Store**
This store contains the identities of other users.  These identities contain the public keys necessary to
encrypt _bundles_ you send to those users, as well as their public signing key so that you can verify
that the sender of the _bundle_ is the expected user identity.

This store file is always encrypted. It is saved as a user bundle.  However, the system read and write
identities are used for the sender and receiver identities. Those identities are located in the
_Local Identities Store_ described above in section **a**.

As previously mentioned, if you choose to encrypt the _Local Identities Store_ with a user-supplied
symmetric key, and then somehow lose or forget the key, not only will you be unable to decrypt the 
_Local Identities Store_, but you will be unable to access the _User Identities Store_ as well. This is
because the system read and write identities are no longer available, which are used to encrypt the
_User Identities Store_.  In this sense, the optional user-supplied symmetric key used to encrypt
the _Local Identities Store_ also protects the _User Identities Store_, since it protects access to
the system read and write keys in the _Local Identities Store_.

Similar to the _User Bundles_, the _User Identities Store_ is written out as a single bundle file and
the same cryptographic techniques are used to encrypt it.  Those are ed25519 and curve25519 for signing and
encryption.

## **IV. Bumblebee Assets**
The following is a list of assets relating to the Bumblebee CLI environment.

### **A. Local Profiles**
Local Profiles are local file paths that store keys and related data, which are used for sending and 
receiving bundles with other Bumblebee users. There may be more than one profile per local installation.

#### **1. Keypair Identity Store**
The KeyPair Identity Store is a single file that contains the private keys used for reading bundles you
receive, as well as signing bundles you send.  A pair of two key sets, one for encryption and one for
signing, defines an identity.  Identities are referenced by a username.  You may have any number of
identities.  The private key components of identities are only stored locally in the KeyPair Identity Store.
The public components are shared with other users.

This store is optionally encrypted with a user supplied symmetric key.  When encrypted, it is always
encrypted as a complete stream on every write.

#### **2. User Identity Store**
The User Identity Store is a single file that contains the public key components for other _Bumblebee_ users. 
Each user identity contains two public keys, one for encrypting secrets and another for validating the
signature provided with the secret.

This file is always encrypted as a complete stream on every write using system keys stored in the
Keypair Identity Store.

### **B. Short-Lived Bundles**
Short-lived bundles define a lifecycle where a message is intended to be shared and then discarded.
Meaning, you create a bundle for another user, provide that bundle to the user in some way,
then the user consumes the bundle.  After it is consumed, it is discarded. 

The primary distinction here is that this bundle has no expectation of long-term scope.  Usually,
these are secrets with a transient life scope.  It might be credentials for a system.  While the credentials
themselves may have a longer life, you do not expect the bundle with that data to be relevant for very long.

### **C. Long Lived Bundles**
Long-lived bundles have an indefinite life expectancy.  They may need to be maintained in an accessible
form for a long time, from days to months to years.  Generally, these would be bundles you create for
your own purposes, specifically use cases involving storage or backup in some form.

_Bumblebee_ provides explicit functionality for this use case in the form of the ***--local-keys***
flag.

The scope of a long-lived bundle infers that the key sets must endure in accessible state for the duration
of the bundle.

## **V. Potential Threats**
The following is a list of potential threats that an attacker would target within the _Bumblebee_
ecosystem.  These are not necessarily unique to _Bumblebee_ and are generally known threats for any
hybrid encryption system similar in nature to _Bumblebee_.

### **A. Attacker gains access to local system and user account data**
These class of threats concern outcomes when an attacker is able to gain access to the local system
that _Bumblebee_ is installed on.  Specifically, they must either gain access to the system's user account
that has installed _Bumblebee_, or they must gain root or system admin level access.  This will differ per
operating system.

If this access is achieved, the follow potential threats are concerns.

### **1. Access to account stores**
The attacker has access to the _Bumblebee_ profile paths that contain both the local private identities,
and the shared public keys of other users.  If they obtain the private keys of the local identities,
they can potentially use them to executive successful MitM (Man in the Middle) attacks on messages that are
sent to the compromised user's account.

In addition, they could use combinations of those keys offline to send messages to those public key entities
that appear to come from you.

Of course, they could just use _Bumblebee_ to send messages to others on your behalf, without having to
directly access the key data itself.

They could also copy the entire _Bumblebee_ profile path and move it to their own system, basically
reproducing your environment.  This would allow them to send and recieve messages on your behalf.

### **2. Change of identity keys**
With this access, they could change keys at will.  The resulting gains from this would be varied in nature.

### **3. Loss of key data**
They could destroy key data, which would prevent you from decrypting messages you receive, or prevent you
from sending messages to other entities as you did before.

You would also lose the ability to decrypt any long-lived bundles.  Those bundles would be rendered
unreadable.

### **B. Attacker destroys or damages local system**
If an attacker gains physical access to your system, they can physically destroy it. In addition,
if they gain remote control in some way, depending on the operating system and other factors, they could
potentially destroy the operating environment in such a way it is no longer usable.  

Both of those scenarios, whether physical or remote, can result in rendering your system unusable.  In such
an event, the following potential threats are concerns.

### **1. Loss of keys**
All of your identities contained on the system be lost.  You would no longer be able to decrypt
messages you receive from other users.  You would also not be able to send messages to the other users.

Any long-lived bundles would no longer be readable.

### **2. Loss of bundles**
The loss of the system would mean any bundles contained in the system would be lost.  This would include
long-lived bundles.  Any bundles not decrypted and accessed previously would be lost.

### **C. Brute force attack of bundle payload's symmetric key**
Each bundle consists of a header and a payload.  The payload is encrypted with a random, complex
symmetric key.  That key is further strengthened with Argon2 and then stored in the header using
asymmetric encryption.  

While retrieving the payload key from the header is assumed not possible without the receiver's private
key, an attacker could executive a brute-force attack on the symmetric key of the payload itself.
If such an attack were successful, it would grant the attacker full access to the original payload
data.

## **VI. Potential Vulnerabilities**

### **A. PKI private/public relationship deconstruction**

### **B. Compromise of _Bumblebee_'s code**

### **B. Compromise of third party libraries**

### **1. Go crypto library**
This includes random functionality and crypto implementations

### **2. NKEYS Libraries**

### **C. Compromise of system's random sources**
These are the system entropy sources utilized by the crypto code

### **D. Compromise of local identities**
Discuss MITM issues if private keys are compromised.

### **E. Compromise of user identities**
Ability to fake signature verification identity

### **F. System crash resulting in loss of local and remote identities**

## **VII. Risk Assessments**

## **VIII. Mitigations**

## **IX. Incident Response Plans**
