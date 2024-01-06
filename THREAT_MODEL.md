# A Threat Model Of The Bumblebee Project

## **I. Document Purpose**
Like all security-related solutions, the _Bumblebee_ project consists of a number of trade-offs,
attempting to balance security objectives with practical software realities.  In light of these trade-offs, 
you may find that _Bumblebee_ is more appropriate for some use-cases and usage scenarios than others.  

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
In the context of _Bumblebee_, a "**Secret**" is simply something you do not want other people or systems
to know.  For example, your username and password for your online banking account is a ***secret***, in
that you definitely do not want other entities to know your banking credentials.  In that sense, any info
you want to keep private is a secret, whether it is credentials, documents, etc.

### **D. Bundles**
Bundles are data streams that store your secrets.  Generally these are files, but may be console outputs,
clipboard data, or pipe data.  Bundles refer specifically to the data sequence formats and constructions
that _Bumblebee_ uses to move secrets around.

## **III. A Description Of The Bumblebee Project**

### **A. Overview Of The CLI**
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

### **B. Cryptographic Techniques And Algorithms Employed**
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

### **A. Attacker Gains Access To Local System And User Account Data**
These class of threats concern outcomes when an attacker is able to gain access to the local system
that _Bumblebee_ is installed on.  Specifically, they must either gain access to the system's user account
that has installed _Bumblebee_, or they must gain root or system admin level access.  This will differ per
operating system.

If this access is achieved, the follow potential threats are concerns.

### **1. Access To Account Stores**
The attacker would have access to the _Bumblebee_ profile paths that contain both the local private identities,
and the shared public keys of other users.  If they obtain the private keys of the local identities,
they could read bundles sent to your local identities by other users. 

_Bumblebee_ currently does not provide any network
functionality or transfer protocols, so in this case, there are no practical MITM (Man In The Middle) concerns.
Even if an attacker were able to gain access to bundles that you have sent to another user, they cannot read
them or reconstruct them into a workable bundle, without the other user's private keys.

However, they could use your identities to sign and send bundles to other users that would appear to come
from you.

Of course, they could just use your local _Bumblebee_ environment to send messages to others on your behalf,
or read bundles send to you, without having to directly access the key data itself.

They could also copy the entire _Bumblebee_ profile path and move it to their own system, basically
reproducing your environment.  This would allow them to send and receive messages on your behalf.

### **2. Change Of Identity Keys**
With this access, they could change keys at will.  The resulting gains from this would be varied in nature.

### **3. Loss Of Identity Data**
They could destroy identity data, which would prevent you from decrypting messages you receive, and/or
prevent you from sending messages to other entities as you did before.

You would also lose the ability to decrypt any long-lived bundles.  Those bundles would be rendered
unreadable.

### **B. Attacker Destroys Or Damages Local System**
If an attacker gains physical access to your system, they can physically destroy it. In addition,
if they gain remote control in some way, depending on the operating system and other factors, they could
potentially destroy the operating environment in such a way it is no longer usable.  

Both of those scenarios, whether physical or remote, can result in rendering your system unusable.  In such
an event, the following potential threats are concerns.

### **1. Loss Of Identities**
All of your identities contained on the system could be lost. You would no longer be able to decrypt
messages you receive from other users.  You would also not be able to send messages to the other users.

Any long-lived bundles would no longer be readable.

### **2. Loss Of Bundles**
The loss of the system would mean any bundles contained in the system would be lost.  This would include
long-lived bundles.  Any bundles not decrypted and accessed previously would be lost.

### **C. Brute Force Attack Of Bundle Payload's Symmetric Key**
Each bundle consists of a header and a payload.  The payload is encrypted with a random, complex
symmetric key.  That key is further strengthened with Argon2 and then stored in the header using
asymmetric encryption.  

While retrieving the payload key from the header is assumed not possible without the receiver's private
key, an attacker could executive a brute-force attack on the symmetric key of the payload itself.
If such an attack were successful, it would grant the attacker full access to the original payload
data.

## **VI. Potential Vulnerabilities**
The following is a list of known vulnerabilities that may apply to _Bumblebee_.  These vulnerabilities
are not necessarily unique to _Bumblebee_, but may be generally applicable to any crypto system that is
similar in nature to _Bumblebee_.

### **A. PKI Private/Public Relationship Deconstruction**
_Bumblebee_ uses ***hybrid crypography***, which means it employs both **asymmetric** and **symmetric**
cryptographic technology to achieve its goals.  The **asymmetric** functionality relies on theories
of key pair systems, where two keys are used to accomplish encrypting and decrypting of data.

In this system, one key is designated as the **private key** and one is designated as the **public key**.
You provide your **public key** to other users so that they can exchange messages and data with you.
However, you are required to keep your **private key** secret. As long as you do not provide your
**private key** to someone else, and your system is not compromised or accessed to obtain the 
**private key** in some way, then your key is secure.

Your public and private keys are linked mathematically in a way that makes it very difficult to derive one
from the other, and is considered not possible with known, current technologies.  Generally,
most of the modern cryptography industry designs systems around this assumption.  

However, if this assumption that keys are not derivable from each other were to be found to be not true,
this would break _Bumblebee_'s entire security context.  Currently, this is not believed to be the case.
However, two scenarios are of possible concern here.

The first scenario is if a very sophisticated technology group were to somehow invent a way to derive a
private key from the public key.  This is not believed to be possible, but if it were to occur, then this would
result in a vulnerability within the asymmetric functionality.

The second involves theoretical abilities of emerging quantum computing technology.  It is possible that
once quantum computing attains some level of ability, it could potentially bypass or defeat specific key pair
associations.  This is a theoretical concern, but one that is being addressed by NIST in its 
"post-quantum security" efforts.

It is not known when quantum computing will obtain such an ability, if ever.

### **B. Compromise Of Future Security**
_Bumblebee_ does not use a networking solution to connect to other users.  All processes occur "offline",
in that you perform them on your local system without the need for a network interaction with the user
to whom you are sending messages and data.

As a result, unlike TLS and similar network protocols, there is no way to establish ephemeral session keys
and states.  This means there is also no mechanism for achieving future secrecy, such as DH exchanges.

So, while the symmetric keys for each bundle are random and are not susceptible to future secrecy concerns,
the asymmetric functionality is susceptible to this.

This means that if an attacker is able to obtain your private key, they will be able to read bundles
sent to you.  It also means that they will be able to read any bundle that has ever been sent to you,
or any bundle that will be sent to you in the future, with that same key set.  This is known as
"future secrecy."

If any entity has collected all of the bundles sent to you, then they somehow obtained your private key,
they could conceivably read all those bundles.

Future technologies in _Bumblebee_ may address this for certain use cases.

### **C. Compromise Of _Bumblebee_'s Code**
_Bumblebee_ is an open source project.  The source code is publicly available.  It would be very difficult
for a programmer with ill intent to inject a compromise into the source code of the project.  There are
controls in place to prevent this.  However, it is worth mentioning in this list, regardless of how
improbable such an event would be.  Nevertheless, if somehow someone was able to do so, this could result
in a number of unpredictable compromises.

### **D. Compromise Of Dependent Libraries**
_Bumblebee_ uses a number of other open source projects to accomplish its goals.  Similar to the
unlikely event with _Bumblebee_'s code, it would be rather difficult to inject a compromise into those
libraries as well, for the same reasons.

Regardless, we will call out those critical dependencies here.

### **1. Go Crypto Library**
_Bumblebee_ is highly dependent on the Go language crypto libraries.  This includes both for the use of
its cryptographic algorithms, and for cross-platform random number input sources.

This library is predominantly controlled by staff at Google, which would make it difficult to compromise
without community awareness.

### **2. NKEYS Libraries**
Synadia maintains the NKEYS packages.  _Bumblebee_ uses the NKEYS library for asymmetric functionality
and some helper functionality.  Synadia uses these packages for their own security requirements in the
NATS ecosystem.  Any compromise here would be known before it is used in the _Bumblebee_ environment.

### **C. Compromise Of System's Random Sources**
The Go crypto packages provide functionality for accessing cryptographically strong random input
sources. These system mechanisms differ per operating system.

If the system's random input source itself is somehow compromised, this could result in weakening certain
cryptographic functions that _Bumblebee_ uses.

### **D. Compromise Of Local Environment And User Account**
If an attacker gains access to your local environment, and specifically, the assets of a user account with 
_Bumblebee_ installed, they could access your _Bumblebee_ profile data. This would include your private
identities and the public identities of other _Bumblebee_ users that you interact with.  

Access to your private identities would allow them to read bundles sent to you, as well as sign data on your
behalf that is sent to other users.  Additionally, access to the public identities of other users would
allow them to send data to those users on your behalf.

### **E. System Crash Resulting In Loss Of Local And Remote Identities**
If your system crashes, you may lose some or all of your _Bumblebee_ environment.  This would include
some or all of your profile data.  If you lose your local identities, you would no longer be able to read
bundles sent to you.  If you had long-lived bundles, such as for backup purposes, you would not be able
to read those anymore.

## **VII. Risk Assessments**
The following are general assessments of impact severity for various compromise scenarios.  Given that
the core purpose of _Bumblebee_ is to exchange data securely, risks are generally related to a scenario
where that data is compromised.  All other risks are operational level, including the impact of time to
restore processes or mitigate the results of some attack.

### **A. Compromise of Private Keys: Low to High**
If someone was able to access your private keys, this would allow them to read bundles sent to you.
The impact would be commensurate to the nature of the data that is sent to you.  This could vary from a low to
high level of impact, depending on the sensitivity and classification of that data.

### **B. Loss of Local Environment Due To Crash Or Attack: Low To High**
If your local system crashes or is destroyed for some reason, it would be very frustrating.  However,
this will not result in a compromise.  To the contrary, it would likely result in loss of data.

If the lost data was of nominal value, the severity would be low.

If the lost data represented mission-critical artifacts, the impact would be commensurate to the value
of that data.

## **VIII. Mitigating and Preventing The Potential of Threats and Vulnerabilities**
Because _Bumblebee_ does not currently use any networking functionality, the majority of prevention
and mitigation really boils down to I.T. best practices and awareness.  Most of the following
recommendations will fall into those categories.

### **A. Preventing Loss Of Environment Using Standard Recoverability Practices**
_Bumblebee_ intentionally uses simple file constructions for your profile data.  All profiles and
identity-related data is stored under a single path in your user's configuration path.  This will differ
per operating system.

The config path for each operating system should be:

    Linux: $HOME/.config/bumblebee
    
    MacOS: $HOME/Library/Application Support/bumblebee
    
    Windows: C:\Users\%USER%\AppData\Roaming\bumblebee

All _Bumblebee_ profile configs are located in that path.  Each profile will be stored in a separate
directory under that root path.  

You can back up all of your identity and config by backing up that directory. If you should have a system
failure or someone should destroy your _Bumblebee_ enviornment, you can simply restore it to a set point
in time.  Be sure to back up your data after any major change to your local _Bumblebee_ environment,
such as after creating a new identity.

Before backing up the directory, you can optionally apply a symmetric password to each profile using the
command **bumblebee set password keypairs**. This is to prevent access to your identities if your backup
is compromised. If you should need to restore it later for use, then you can remove the password at that
time.  There is a _Bumblebee_ future feature planned to provide functionality for encrypted backups of your
config and environment data.

If you have any long-lived bundles, you will want to back those up as well, in case of a loss of system
data or functionality.

### **B. Implement System Access Policies And Configurations**
For any system with _Bumblebee_ installed on it, you will want to implement access control measures.
This means your system should require a full login that at least includes a username and password to
access your system. You can also implement multifactor auth for stronger security requirements if you need.

### **C. Apply Profile Passwords To Encrypt Your Local Identities**
If your environment is high risk for access, or you are bundling very sensitive data, you can apply
a password to your profiles using the command **bumblebee set password keypairs**.  This will encrypt
your local identity store with strong symmetric encryption using the supplied password.  This also
restricts access to the user identity store, because the system keys are now encrypted in the
local identity store.

When you apply a password in this way, _Bumblebee_ will require you to enter the password every time you run
_Bumblebee_.  There is an option to supply the password via an environment variable, if that is a secure
and viable option for you.

Be sure to securely store that password in some way, unless you don't need to for some reason.
If you forget it or lose it in some way, you will no longer be able to access that profile and your
identities that are located there.

There may be a future feature added to _Bumblebee_ to support hardware-based keys, such as dongles or
storing keys on removable drives.

## **IX. Incident Response Plan**
The primary concern relating to a _Bumblebee_ compromise is if an attacker gains access to your
private identities stored in your local environment, or just access to the _Bumblebee_ environment
itself.

In either case, depending on the specifics of the compromise, this allows an attacker to read bundles
sent to you, as well as send bundles to other users on your behalf.

If a compromise of this nature occurs, you may respond differently depending on whether you had used
a password to encrypt your local identities.  If you were not using a password, you definitely want to
process with the recommended steps below.

If you were using a password, then you want to consider how strong the password is and how sensitive
are the secrets you are protecting?

If your password was weak, something like "abc123" or your birthday, then proceed with the steps listed.

If your password was strong, then you may want to consider the fact that the only known successful
attack on your identity data is going to be a brute-force attack on the symmetric key.

If the data you are protecting is not critically sensitive, and you were using a strong password for
your local identity store, you MIGHT consider this a low level compromise and do nothin other than
**Step A** below.

However, it is recommended that you want to err on the side of considering this a high-level compromise
and execute the steps listed below.

Whatever you decide, you always want to follow **Step A** at least.

### **A. Resolve The Vulnerability Used To Compromise Your Environment**
The first step is to fix whatever vulnerability was used to gain access to your _Bumblebee_ environment.
Before you go through the effort to reset your identities, you want to make sure that they can't easily
compromise your environment again.

### **B. Recover Any Secrets Captured In Long-lived Bundles**
If you have stored secrets in long-lived bundles, you will want to retrieve those secrets BEFORE you
deprecate the compromised identities.  For example, if you use the **--local-keys** option to encrypt
files for backups, you will want to recover the data stored in those files.  If you have a strong access
control to your files, you MIGHT consider keeping a backup of the compromised keys for future access.
However, now that those identities are compromised, it's best to stop using those identities going forward.

### **C. Deprecate The Compromised Identities By Initializing New Identities**
To deprecate the current identities and generate new ones, you would follow the same steps as when
you first setup _Bumblebee_.  This would be to run the **bumblebee init** command.

Since you already have profiles created, _Bumblebee_ will ask you if you wish to remove them.  Select
_YES_ and _Bumblebee_ will remove all current profiles and create a new, clean environment.

Just like when you first setup _Bumblebee_, you will now have an environment with just a single
profile named **default**.  Completely new keys are generated in the new environment and are safe to
use.

At this point, you will need to import other users as you did before.

If you have more than a default profile, you will need to re-create those now, which will also
be generated with new identities.

### **D. Inform Other Users About Your Compromise And Provide Them With Your New Public Keys**
You will want to let any associated users know that your identity has been deleted and re-created.
They should no longer use your prior public keys.

There are several ways to do this, but the easiest for them is to just delete the current reference to
your public account and re-import your data.

They can remove your current user info with this command...

    bumblebee remove user <username>

Then, you will provide them with an export of any new identities you created for them to use.  This is
the command for exporting your private identity...

    bumblebee export user [name] --from-keypair

Then they will import your data with this command...

    bumblebee import --input-file [filename]

At this point, the other users can safely send and receive bundles with you again.
