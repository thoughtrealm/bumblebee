# Bumblebee Technical Details

## Overview of _Bumblebee_ functionality
_Bumblebee_ uses a ***hybrid cryptography*** approach.  This means it uses both ***asymmetric*** and
***symmetric*** cryptographic functionality.  This is a common, well-known technique utilized by numerous
security and messaging products.

The _Bumblebee_ CLI (Command Line Interface) maintains a local environment, which hosts stores for
your personal and system identities, as well as public keys for other users.  You can share the public keys
using the _Bumblebee_ **import** and **export** commands.  It is also possible that future functionality may
provide the sharing functionality via a network service, so that you don't have to export and import them.

Using these identities and user references, _Bumblebee_ allows you to create an encrypted ***Bundle*** for
delivering to a specific _Bumblebee_ user. Using _Bumblebee_, the receiving user is able to decrypt
the ***Bundle***, as well as validate the sender's identity.

_Bundles_ can be shared in several forms, including binary files and text safe forms.  Text safe forms
can be shared using text messages, Slack posts, text documents, clipboard storage, etc.

## General description of Bumblebee's cryptographic approach
When you set up and initialize _Bumblebee_, a **default** profile is created.  Additional profiles can also
be created, all of which provide the same functionality within the security context of that profile.

The security context is compromised of local identities and external user public keys.

To accomplish this, a profile contains two separate ***stores***.

1. **The Store of Local Key Pair Identities**<br>
    One store is referred to as the ***Keypair Store***.  This store contains your keypair identities that
    are referenced by a name, with the possible exception of the **default** identity.  It can store any
    number of identities.  During initialization, a **default** identity is created.  However, you may 
    create an identity named "home", or "work", or whatever.<br>
    
    Each of these identities consists of two 25519 key sets.  One is a curve25519 for encrypting data, while the
    other is an ed25519 key set for signing data.<br>

    Two additional system key sets are created for each profile named ***keystore_read*** and ***keystore_write***.
    These system keys are used for encrypting the ***User Store***.  They can also be used for encrypting personal
    assets that are not intended for delivery to another user, such as encrypting files for backup, etc.<br>

    The ***Key Pair Store*** is stored in your local config path.  It can be optionally encrypted with a
    symmetric key, which you would provide in order to decrypt the store's file stream.  If the store
    is encrypted with a symmetric key, it is requested during launch when running a Bumblebee command.  It
    can also be provided using a local environment variable.<br>

    This store contains all of your private keys. If this store is lost in some way, you will not be able to 
    decrypt assets that have been encrypted with these keys.<br>

    These identities are stored in a single file.<br>
    
2. **The Store of User Identities**<br>
    The other store is called the ***User Store***.  This stores the identities of other users that you
    share secrets with. The user is stored based on names, which can be handles, email addresses,
    LDAP account names, etc.<br>

    Each user identity consists of two public keys.  One is the user's curve25519 cipher public key.  
    The other is the user's ed25519 signing public key.  These keys are used when creating bundles, as well
    as when validating _bundle_ signatures.<br>

    The User Store is stored in a single file.<br>

The profile consists of these two store files, plus some metadata stored in YAML files.

To back up these profiles, their stores and all their data, you only need to archive and/or copy the
config path to some backup path or container.

## Cryptographic code used and the related origins
1. Bumblebee is written in the **Go** language.  All functionality is contained in **Go** code.<br><br>

2. The Asymmetric functionality is provided by the **NATS** **NKEYS** packages.<br><br>

   **NATS** is an open source, cloud scale messaging server and related ecosystem.  The project is found at
   https://nats.io/. <br><br>

   The **NKEYS** packages provide the distributed security mechanisms for **NATS**.  They are found at
   https://github.com/nats-io/nkeys.  <br><br>

   **NKEYS** provides functionality for ed25519 signing and verification. 
   **XKEYS** is a package within **NKEYS** that provides curve25519 encryption.
   These packages provide a _SALT/NaCL_ compatible library that wraps the corresponding **Go** implementations,
   providing easier key handling formats and other helper logic.<br><br>

3. The symmetric functionality utilizes the _XChacha20-poly1305_ cipher. Key strengthening utilizes
   _Argon2_. All of these are provided by the **Go** crypto packages.<br><br>

4. The random sequence generations are done using **Go**'s _crypto/rand_ package, which provides crypto strength
   random functionality on all supported platforms.<br><br>

## Creating a Combined Bundle
A combined bundle is where the bundle's header and payload (data) are emitted to the same stream.

The following describes the logic for creating a combined bundle:
1. Create a random symmetric key.
2. Sign a random construction with the sender's private signing key.
3. Store the symmetric key, signature stream and some related metadata in a bundle header.
4. Encrypt the header using curve25519 and the receiver's public key.
5. Emit the header to the stream. 
6. Strengthen the symmetric key with Argon2.
7. Using the strengthened symmetric key, iterate and encrypt the secret message in ~32k chunks using
XChacha20-poly1305 and an AD value.  Emit each encrypted chunk to the output stream in order.

## Opening a Combined Bundle
The following describes the logic for reading a combined bundle:
1. Extract the bundle header from the stream.
2. Decrypt bundle header using the receiver's private key.
3. Extract and validate that the signature stream in the header was signed by the expected sender.
4. Extract the symmetric key and strengthen it with Argon2.
5. Using the strengthened key, iterate and decrypt the bundle payload data in ~32k chunks using
XChacha20-poly1305, using the AD value for data validation, and then write each decrypted chunk to
the output stream. 

## Minor Logic Differences for Split Stream Outputs
The logic differences between combined and split stream support are very minor.

When creating the bundle, terminate the stream after writing out the header.  Initialize the data stream
and emit the encrypted data to that stream.

When reading the data, read the header stream and process it as you would the combined stream.  Once
you have the key in-hand and have validated the header info, read and decrypt the separate payload stream. 

## Bundle stream diagram
[This diagram](docs/StreamCompositionOfBundles.pdf) describes the layout of the combined and split stream bundles. 

## Describing the Bundle With an Analogy of "**Two Stages of Locked Boxes**"
To better understand the bundling process, as well as it's strengths and weaknesses of the bundles, we
attempt to describe it using an alaogy of "_Two Stages of Locked Boxes_."  

This analogy is surely not unique and is likely referred to by others using some much more clever
sounding name.  Nevertheless, it is the name we will use for this analogy when describing this approach.

The motivation for this approach is because, in our offline scenario, we must encrypt the data without
the availability of an active network session.  As a result, there is no active exchange mechanism that
can allow us to derive ephemeral, mutually agreed upon session keys via DH exchange mechanisms, nor other
supporting behaviors when engaging over an active session.

One option we might choose would be to require User 1 to provide a symmetric key for encrypting the data,
which they would then be responsible for sharing both the data and the key with User 2. User 2 could
then decrypt the data. However, this is overly complicated and would depend on too many user-focused processes.

Instead, we use Hybrid Encryption, taking advantage of the strengths of both Asymmetric and Symmetric
cryptography.  This would be somewhat similar to the approach used by SSL/TLS, but without the support
of DH agreements, and naturally without those benefits as well.

We first generate a random, strong symmetric key which we do not reveal to the user directly.  We use that
symmetric key to encrypt the secret.

In a slightly over simplified description, we then we encrypt that key and some other elements
(salt, signature, etc) using asymmetric cryptography.  The asymmetric cryptography requires sharing only
the public keys, which can be shared safely.  This prevents having to know and manage the sharing of the
symmetric key itself.  Then we encrypt the secret data using symmetric cryptography and the symmetric key.

To understand this, let's say that...
 
- And there are two users, Alice and Bob.

- You have two boxes, Box 1 and Box 2.

- Box 1 requires a single key to open it, which we'll call Key 1.  There is only one known copy of Key 1 and
Box 1 is considered indestructible.  Once locked, it can only be opened using Key 1.

- Box 2 requires two keys, key A and key B. It is locked with Key A and can only be opened with Key B.
Box 2 is also considered indestructible.  Once it is locked with Key A, it can only be opened using Key B.

- There are an unknown number of copies of Key A, from one to possibly many.  They may be in the possession
of friends or enemies of Alice and Bob. Regardless, none of them can open box 2 with their copies of Key A.

- There is only one known Key B in existence, which is in Bob's possession.

- Alice takes her secret, puts it in Box 1 and locks Box 1 with Key 1. 

- She then takes Key 1 and places it in Box 2.

- She also includes a secret note, which she signs in her own handwriting, and puts it in Box 2.
The handwriting will confirm for Bob that Alice is indeed the one who sent Box 2.

- She then locks Box 2 with Key A.  

- Now that Box 2 is locked, neither she nor anyone other than Bob can open Box 2.
Only Bob can open Box 2 with his Key B.

- Alice sends both boxes to Bob using any mechanism she wishes.
She can hand-deliver them, mail them, leave them out for Bob to come by and pickup, etc.

- If an enemy should acquire either of the boxes, they are unable to extract the contents of
either without Key B. In that regard, Alice's secret is safe.

- Assuming Bob takes possession of both boxes, he uses his Key B to open Box 2.

- He examines the secret note at this point, and confirms it is signed with Alice's handwriting.
If it is not her handwriting, he discards Box 1 and does not open it.

- Otherwise, he then removes Key 1 and opens Box 1, extracting Alice's secret.

This is how that analogy correlates technically to the Bumblebee approach:

- We generate a random, strengthened key and encrypt the secret data with it using symmetric cryptography.
This is Box 1 and Key 1 in our analogy.

- We build a separate data structure called the Header.  In this structure, we put various things,
including the symmetric key (Key 1 in our analogy) and a data element signed with the sender's private
signing key (the secret note).  We encrypt this information with asymmetric cryptography, using the
receiver's public Key.  This is Box 2 and Key A in our analogy.

- Both are delivered to the receiver, who unlocks the asymmetric structure with their private key
(Box 2, Key B).

- Bumblebee uses the stored signature to confirm that the sender is who we expected (the handwriting on
the secret note). 

- If the signature is verified, _Bumblebee_ then extracts the symmetric key (Key 1) and decrypts the secret
data (Box 1).

This process does not require the sender or the receiver to manage any of the cryptographic elements mentioned in the process, outside of sharing public keys in some way.  The sharing of public keys is a one-time process, unless they are changed in the future for some reason.

## Extending The "Two Stages of Locked Boxes" Analogy With Two Shipments
Now, let's extend the analogy a little bit.  Perhaps, Alice is not comfortable with the security of putting
Key 1 in Box 2 and providing both boxes at the same time.  Basically, her concern is that transporting Key 1,
which unlocks Box 1, along with Box 1, is inherently insecure, even if it is in Box 2, which is impenetrable.
It just makes her uncomfortable.

So, Alice ships the two boxes separately.  She sends Box 1 via USPS and Box 2 via FedEx.  Assume that they
take different routes and arrive at different times. Alice feels better about this, because Key 1 is never
transported in the presence of Box 1.  Also, without both boxes, the secret cannot be accessed.

Once Bob has both boxes in his possession, he is able to access her secret, as previously described.

Relating this to the Bumblebee bundling process, when you bundle a secret, you can choose whether to output
the bundle to a single "combined stream" or two "split streams." This is effectively what Alice is dealing
with in this scenario extension.

You can use set the "Bundle Type" by using the **--bundle-type** flag (or **-b** shortcut).
The options are ***combined*** or ***split***.

With ***combined***, the output is emitted to a single stream.  The default extension for this is ***.bcomb**
for "Bumblebee combined stream" format. The extension can be overridden by providing an explicit file name
using the **--output-file** flag (or **-y** shortcut). When using the combined output encoding, the length
of the header is emitted to the start of the output stream, followed by the header data, which is then
followed by the payload data.  When opened by the receiver, the header and payload are extracted from the
combined stream and processed to provide the unencrypted secret.

With ***split***, two separate output entities are created; one for the header and one for the payload.
The header will have a default extension of ***\*.bhdr*** and the payload will have a default extension
of ***\*.bdata***.

You may deliver the two separate files any way you wish. When opening the bundle, you specify the split
bundle type, and Bumblebee will process the two components accordingly to create the unencrypted output.

It is a bit more work to provide the two separate artifacts using different transport paths.  Otherwise, both
combined and split encodings are functionally identical.

_**Note**: While the combined stream should be sufficiently secure for our needs, if you are concerned that
a weakness could be exploited with the combined approach, you can choose to use split streams.  Perhaps
asymmetric key associations are rendered ineffective or insecure due to some emergent tech (e.g. quantum
advances), then using split streams will mitigate that concern to some degree.  Of course, in that event,
most modern cryptographic systems will be rendered ineffective as well.

The soon coming public reveal of post-quantum solutions will shed more light on this.  Once a post-quantum
solution is accepted by the community, perhaps _Bumblebee_ would be updated to use those N.I.S.T.
recommended algorithms accordingly.

## A General Technical Description Of The **BUNDLE** Process
The _Bundle_ process receives an input byte sequence (secret) and outputs it as an encrypted byte sequence
compromised of two parts:
1. A Bundle Header that is encrypted with curve25519
2. A Bundle Payload that is encrypted with XChacha20-Poly1305

The header contains various elements of details and metadata, and the payload contains the secret data.

The following items are included in the header as of Bumblebee release 0.1.0:

<pre>
SymmetricKey     : A random value used to encrypt the payload using XChacha20-Poly1305

Salt             : A random value provided for the payload encryption

InputSource      : TYPE BundleInputSource- the source of the data provided for bundling

CreateDate       : The date the bundle was created in RFC3339

OriginalFileName : The file name of the source file, IF the source was a file

OriginalFileDate : The date stamp of the source file in RFC3339, IF the source was a file

ToName           : The name used to identity the receiver’s public key

FromName         : The name used to identity the sending keypair that encrypted the bundle

SenderSig        : The RandomSignatureData

HdrVer           : The version of the bee functionality that built the header

PayloadVer       : The version of the bee functionality that built the payload
</pre>

When bundling the input, the header is first populated with the following values:

<pre>
- SymmetricKey is set to a random 32-byte sequence
  
- Salt is set to a random 32-byte sequence
  
- SenderSig is type RandomSignatureData, which is initialized with a random 32-byte sequence
  that is signed using ed25519 and the Sender's Private Key from their ed25519 (signing) keypair.
  Both the input random sequence and the signature output are stored in the header.  

- All the remaining metadata values are populated as needed
</pre>

_***Note***: In general practice, we would possibly use a different signing approach, such as hashing
the payload data and signing that value.  In our bundle process flow, we have not encrypted the secret at
this point, and we may not even have the secret in hand yet, given that it may be a streaming input.  So,
we do not know enough about the secret to sign something that is a computational function of the secret,
whether before or after encryption.  Therefore, with the Bumblebee bundle flow, we generate a random nonce
for each bundle and sign that value instead.  This is also encrypted as part of the header bytes, so it is
only available to the receiver.  This process is more for efficiency, so we do not have to move around in
the output stream by going back to the header sequence in order to update the header data after payload
emission.  While possibly unconventional, this approach should be quite sufficient for our signing and
sender validation requirements._

The header itself is then encrypted using the **XKEYS** (curve25519) **SEAL** functionality. The **SEAL**
functionality uses the receiver's public Key from their curve25519 keypair.

The header is then written to the output stream.

Once the encrypted header is written to the stream, then the payload is encrypted using the previously derived Salt and SymmetricKey.  While the random SymmetricKey is potentially a strong one-time sequence, it is still strengthened using Argon2. This is to mitigate any weak random sequences that might be generated.

The payload is encrypted using XChacha20-Poly1305, which is a streaming cipher and provides for the output of
large payload streams.  The output encryption is performed in sealed chunks of 32,000 bytes.  Each chunk
will result in a small increase in output size, due to nonce and AEAD overhead, so the resulting output
stream will be slightly larger than the input stream.

## A Pseudocode Description of the **BUNDLE** Process
Given these values:
<pre>
    Key<sub>payload</sub> is a random symmetric key for the payload data stream
    
    Salt<sub>payload</sub> is a random 32-byte salt input for the Argon2 permutation
    Key<sub>derived</sub> is an Argon2 permutation of Key<sub>payload</sub>, which is used to encrypt the
    payload with XChacha20-Poly1305
    
    PUB<sub>receiver</sub> is the curve25519 public key for the receiver
    
    PK<sub>sign-sender</sub> is the ed25519 private signing key for the sender
    
    Salt<sub>sign</sub> is the random input for the signing sequence
    
    Signature is the signed sequence stored in the header
    
    Header<sub>plain</sub> is a bundle header structure as described above
    
    Header<sub>encrypted</sub> is the encrypted Header<sub>plain</sub>
    
    Header<sub>length</sub> is the int16u length of Header<sub>encrypted</sub>
    
    Cipher is an initialized XChacha20-Poly1305 encrypter
    
    ADConst is a value for Cipher's AD (Associated Data)
    
    Secret<sub>input</sub> is the provided secret to encrypt
    
    Secret<sub>encrypted</sub> is the encrypted form of Secret<sub>input</sub>
</pre>

This is the Bundle process:

<pre>
    Salt<sub>payloadM</sub> <= Random[32]
    Key<sub>payload</sub> <= Random[32]
    Header<sub>plain</sub> <= New Header[Key<sub>payload</sub>, Salt<sub>payload</sub>]

    Header<sub>plain</sub>::Salt<sub>sign</sub>	<= Random[32]
    Header<sub>plain</sub>::Signature <= ed25519.Sign[Salt<sub>sign</sub>, PK<sub>sign-sender</sub>]

    Header<sub>plain</sub>::[Metadata] 	<= Set Values[Metadata]
    Header<sub>encrypted</sub> <= curve25519::Seal[Header<sub>plain</sub>, PUB<sub>receiver</sub>]
    Header<sub>length</sub>	<= length(Header<sub>encrypted</sub>)

    WriteToStream[int16uWithFixedEndian[Header<sub>length</sub>]]
    WriteToStream[Header<sub>encrypted</sub>]

    Key<sub>derived</sub> <= Argon2[Key<sub>payload</sub>, Salt<sub>payload</sub>, [time/mem/threads]]

    Cipher <= XChacha20-Poly1305::Init[Key<sub>derived</sub>]

    Iterate Cipher::[Secret<sub>input</sub>]
        Secret<sub>input</sub>[blockX] <= ReadFromStream[blockSize]
        Secret<sub>encrypted</sub>[blockX] <= Cipher::Encrypt[Secret<sub>input</sub>[blockX], ADConst]
        WriteToStream[Secret<sub>encrypted</sub>[blockX]]
</pre>

## A General Technical Description Of The **OPEN** Process
The _Open_ process receives the encrypted bundle byte sequence and outputs the decrypted, restored secret data.

The Bundle may be in a split or combined form.  Please refer to “A General Description of the BUNDLE Process” for structural details of split vs combined bundles, as well as the bundle header composition.  For this description, we are more concerned with explaining the logic of reversing the bundle data into the original secret.

Regardless of split vs. combined source types, the following process is the same:

- Read the header data<br><br>
      
- Decrypt the header data using the receiver’s private key from their curve25519 (cipher) keypair.
This uses the ***XKEYS*** (curve25519) **Open** functionality.<br><br>
  
- Extract the signing data from the decrypted header.  Confirm that the signature matches the expected
sender’s signature using the specified “from” user’s public key from their ed25519 (signing) keypair.
If the validation fails, abort the process or possibly request permission to proceed.<br><br>
  
- Extract the payload salt and key from the decrypted header.  Then derive the actual payload key using Argon2.<br><br>
  
- Read and decode the payload, while emitting the decoded data to the requested output target and encoding.
The payload is decrypted using the XChacha20-Poly1305 symmetric cipher.<br><br>
  
- If the original source and the output targets are both files, use the original file details for date
and naming, unless the user has supplied explicit target details.  Future implementations may include other
file or target properties that might need to be applied as well.<br><br>

## A Pseudocode Description of the **OPEN** Process
Given these values:
<pre>
    Key<sub>payload</sub> is the symmetric key for the payload data stream
    
    Salt<sub>payload</sub> is the salt input for the Argon2 permutation
    
    Key<sub>derived</sub> is an Argon2 permutation of Key<sub>payload</sub> used to decrypt the payload
    using XChacha20-Poly1305
    
    PUB<sub>sign-sender</sub> is the ed25519 public key for the sender’s signing keypair
    
    PK<sub>receiver</sub> is the curve25519 private key for the receiver
    
    Salt<sub>sign</sub> is input for the signing sequence
    
    Signature is the signed sequence stored in the header
    
    Header<sub>decrypted</sub> is the header structure from decrypting Header<sub>encrypted</sub>
    
    Header<sub>encrypted</sub> is the encrypted input header structure
    
    Header<sub>length</sub>	is the int16u length of the header bytes in the input stream
    
    Cipher is an initialized XChacha20-Poly1305 decrypter
    
    ADConst is a value for Cipher's AD (Associated Data)
    
    Payload<sub>encrypted</sub> is the encrypted input payload data
    
    Payload<sub>decrypted</sub> is the decrypted output of Payload<sub>encrypted</sub>
</pre>

This is the OPEN process:
<pre>
    Header<sub>length</sub> <= ReadFromStream[int16uWithFixedEndian]
    Header<sub>encrypted</sub> <= ReadFromStream[Header<sub>length</sub>]
    Header<sub>decrypted</sub> <= curve25519::Open[Header<sub>encrypted</sub>, PK<sub>receiver</sub>]
    
    Salt<sub>payload</sub> <= Header<sub>decrypted</sub>::Salt<sub>payload</sub>
    Key<sub>payload</sub> <= Header<sub>decrypted</sub>::Key<sub>payload</sub>
    Metadata <= Header<sub>decrypted</sub>::Metadata
    
    Key<sub>derived</sub> <= Argon2[Key<sub>payload</sub>, Salt<sub>payload</sub>, [time/mem/threads]]
    
    Cipher <= XChacha20-Poly1305::Init[Key<sub>derived</sub>]
    
    Iterate Cipher::[Payload<sub>encrypted</sub>]
        Payload<sub>encrypted</sub>[blockX] <= ReadFromStream[blockSize]
        Payload<sub>decrypted</sub>[blockX] <= Cipher::Decrypt[Payload<sub>encrypted</sub>[blockX], ADConst]
        WriteToStream[Payload<sub>decrypted</sub>[blockX]]
</pre>