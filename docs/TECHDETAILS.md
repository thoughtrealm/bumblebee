# Technical Details

### The "Two Stages of Locked Boxes" Approach
A Bundle is what I refer to as the encrypted output construction created by Bumblebee.

I refer to the Bumblebee bundling approach as "Two Stages of Locked Boxes."  I would imagine this analogy
is not unique and is likely referred to by some other very clever and cryptographer-ey sounding name.  
Nevertheless, it is the name of the analogy that I use for describing this approach.

The motivation for this approach is the fact that, in our offline scenario, we must encrypt
data without the availability of an active network session.  As a result, there is no active exchange mechanism that
can allow us to derive ephemeral, agreed upon session keys via a DH exchange mechanism.  

A complex solution would require User 1 to provide a symmetric key for encrypting the data, which they would then be 
responsible for sharing both the data and the key with User 2, so that User 2 could decrypt the data.

Instead, we use both Asymmetric and Symmetric crypto.  This would be similar to SSL/TLS, but without the
benefit of DH agreements, and naturally without some of the benefits as well.  

We first generate a random, strong symmetric key which we do not reveal to the user directly.
We use that symmetric key to encrypt the secret.

In a slightly over simplified description, we then we encrypt that key and some other elements (salt, signature, etc)
using asymmetric crypto.  The asymmetric crypto uses the public keys which can be shared safely.  This prevents
having to know and manage the sharing of the symmetric key itself.

To understand this...

* You have two boxes, Box 1 and Box 2.  

* And there are two users, Alice and Bob.
 
* Box 1 requires a single key to open it, which we'll call Key 1.  There is only one copy of Key 1.  Box 1
is as close to indestructible as possible.  Once locked, it can only be opened using Key 1.
 
* Box 2 requires two keys, key A and key B. It is locked with Key A and can only be opened with Key B.
Box 2 is also as close to indestructible as possible.  Once it is locked with Key A, it can only be opened using Key B.
 
* There is possibly some unknown number of copies of Key A, from one to possibly many.  They may be in the possession of 
friends or enemies of Alice and Bob. Regardless, none of them can open box 2 with their copies of Key A.
 
* There is only one Key B in existence, which is in Bob's possession.
 
* Alice takes her secret, puts it in Box 1 and locks Box 1 with Key 1. 
 
* She then takes Key 1 and places it in Box 2.

* She also includes a secret note, which she signs in her own handwriting, and puts it in Box 2.  The handwriting
will confirm for Bob that Alice is the one who sent Box 2.

* She then locks Box 2 with Key A.  
 
* Now that Box 2 is locked, neither she nor anyone else can open Box 2. Only Bob can open Box 2 with his Key B.
 
* Alice sends both boxes to Bob using any mechanism she wishes.  She can hand-deliver them, mail them, leave them
out for Bob to come by and pickup, etc.
 
* If an enemy should acquire either of the boxes, they are unable to extract the contents of either without Key B.
In that regard, Alice's secret is safe.
 
* Assuming Bob takes possession of both boxes, he uses his Key B to open Box 2.

* He examines the secret note at this point, and confirms it is signed with Alice's handwriting.  If it is not
her handwriting, he discards Box 1 and does not open it.
 
* Otherwise, he then removes Key 1 and opens Box 1, extracting Alice's secret.

Here's a generalized description of how this analogy correlates to the Bumblebee approach technically:

* We generate a random, strengthened key and encrypt the secret data with it using symmetric crypto.  This is Box 1
and Key 1 in our analogy.

* We build a separate data structure called the Header.  In this structure, we put various things, including
the symmetric key (Key 1 in our analogy) and a data element signed with the sender's private signing key 
(the secret note).  We encrypt this information with asymmetric crypto, using the receiver's Public Key.  
This is Box 2 and Key A in our analogy.
 
* Both are delivered to the receiver, who unlocks the asymmetric structure with their private key (Box 2, Key B).
 
* Bumblebee uses the stored signature to affirm the sender is who we expected (the handwriting on the secret note). 
 
* If the signature is verified, Bumblebee then extracts the symmetric key (Key 1) and decrypts the secret data (Box 1).

This process does not require the sender or the receiver to manage any of the cryptographic elements mentioned in 
the process, outside of sharing public keys in some way.  The sharing of public keys is a one-time process, unless 
they are changed in the future for some reason.

### Details Of The Bundle Process Flow
The Bundle process receives an input byte sequence and outputs a byte sequence compromised of two parts:

- A Bundle Header
- A Bundle Payload

The header contains various elements of details and metadata while the payload contains the input data.  

The following items are included in the header as of Bumblebee release 0.1.0:
```
	// SymmetricKey is a random value used to encrypt the payload using Chacha20/Poly1305
	SymmetricKey []byte
	
	// Salt is a random value provided for the payload encryption
	Salt []byte
	
	// InputSource records the source type of the data provided for bundling
	InputSource      BundleInputSource
	
	// The date the bundle was created
	CreateDate       string // RFC3339
	
	// OriginalFileName records the file name of the source file, IF the source was a file
	OriginalFileName string
	
	// OriginalFileData records the date stamp of the source file, IF the source was a file
	OriginalFileDate string // RFC3339
	
	// ToName indicates the name used to identity the User public keys in the keystore
	ToName           string
	
	// FromName indicates the name used to identity the keypair set that encrypted the bundle
	FromName         string
	
	// SenderSig contains the RandomSignatureData 
	SenderSig        []byte
	
	// HdrVer identifies the version of the bee functionality that built the hdr
	HdrVer           string
	
	// PayloadVer identifies the version of the bee functionality that builtthe payload
	PayloadVer       string
```

When bundling the input, the header is first populated with the following values:
- SymmetricKey is set to a random 32 byte sequence
- Salt is set to a random 32 byte sequence
- The SenderSig is initialized with a random 32 byte sequence that is signed using ed25519 and the 
Sender's Private Key from their ed25519 (signing) keypair.  Both the random sequence and the signature output
are stored in the header.
- All the remaining metadata values are populated as necessary

The header itself is then encrypted using the NKEYS XKEYs (curve25519) SEAL functionality. 
The SEAL functionality uses the receiver's Public Key from their curve25519 (cipher) keypair.

The header is then emitted to the output stream.

Once the encrypted header is emitted, then the payload is encrypted using the previously derived
Salt and SymmetricKey.  While the random SymmetricKey is potentially a strong one-time sequence,
it is still strengthened using Argon2.  This is to mitigate any weak random sequences that might be
generated.

The payload is encrypted using XChacha20/Poly1305, which is a streaming cipher and
supports the output of large payload streams.  The output encryption is performed in
sealed chunks of 32,000 bytes.  Each chunk will result in a small increase in output size,
due to nonce and AEAD overhead, so the resulting output stream will be slightly larger than the input stream. 

### A technical flow of the Bundle process...


Let Key<sub>payload</sub> be a random symmetric key for the payload data stream<br/>
Let Key<sub>derived</sub> be an Argon2 permutation of Key<sub>payload</sub>  
Let Salt<sub>payload</sub> be a random 32-byte salt for the payload data stream<br/>

Let PUB<sub>receiver</sub> be the curve25519 public key for the receiver<br/>
Let PK<sub>sign-sender</sub> be the ed25519 private signing key for the sender<br/>
Let Salt<sub>sign</sub> represent the random sequence for signing the header<br/>
Let Signature represent the signed sequence stored in the header<br/>

Let Header<sub>plain</sub> represent a bundle header structure as described above<br/>
Let Header<sub>encrypted</sub> represent the encrypted bundle header<br/>

Let Cipher be an initialized XChacha20-Poly1305 encryptor<br/>
Let ADConst be a value for Cipher's Associated Data<br/>
Let Secret<sub>input</sub> represent the provided secret to encrypt<br/>
Let Secret<sub>encrypted</sub> represent the encrypted form of the secret<br/>

Salt<sub>payload</sub> <= Rand[32]<br/>
Key<sub>payload</sub> <= Rand[32]<br/>

Header<sub>plain</sub> <= New[Key<sub>payload</sub>, Salt<sub>payload</sub>]<br/>
Header<sub>plain</sub>::Salt<sub>sign</sub> <= Rand[32]<br/>
Header<sub>plain</sub>::Signature <= ed25519.Sign[Salt<sub>sign</sub>, PK<sub>sign-sender</sub>]<br/>
Header<sub>plain</sub>::[Metadata] <= Values[Metadata]<br/>

Header<sub>encrypted</sub> <= curve25519[Bundle<sub>plain</sub>, PUB<sub>receiver</sub>]<br/>

WriteToStream[int16uWithFixedEndian[length(Header<sub>encrypted/sub>)]]<br/>
WriteToStream[Header<sub>encrypted</sub>]<br/>

Key<sub>derived</sub> <= Argon2[Key<sub>payload</sub>, Salt<sub>payload</sub>, time/mem/threads]<br/>
Cipher <= XChacha20-Poly1305::Init[Key<sub>derived</sub>]<br/>

Iterate Cipher::[Secret<sub>input</sub>]<br/>
&nbsp;&nbsp;&nbsp;Cipher::Encrypt[Secret<sub>input</sub>[blockX], ADConst] => Secret<sub>encrypted</sub>[blockX]<br/>
&nbsp;&nbsp;&nbsp;WriteToStream[Secret<sub>encrypted</sub>[blockX]]<br/>