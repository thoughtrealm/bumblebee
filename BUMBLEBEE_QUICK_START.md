# Bumblebee Quick Start Guide
#### _Last updated: Dec 9, 2023_

This document will provide a set of steps that will guide you through installing Bumblebee,
sharing keys with another user, and using a few methods for sharing secrets with them.

This document is focused on simple use cases that will show you some basic functionality in Bumblebee.
While there are a number of additional features and options in Bumblebee, this document will not go into those details.

These steps were tested on Debian ARM64.  However, they will work fine on Windows and Mac as well.

## Sending a file to another user
We will go through a simple scenario of sending a file to another user.  There are other ways to share
secrets without files, but we will focus on sharing files first.

Before we can do so, we will need to install Bumblebee, as well as set up a user that
we wish to share secrets with.

After the setup is done, we will walk through steps to demonstrate the general pattern for
sharing files with other Bumblebee users.  

The basic pattern is as follows:
<pre>
1. Run the <b>bundle</b> command and create the encrypted <em>bundle</em> for another user.  
2. Supply the encrypted <em>bundle</em> to that user.
3. The other user then decrypts the <em>bundle</em> with their private keys using the <b>open</b> command. 
4. <em>Bumblebee</em> validates that your identity was indeed the sending identity when it opens the <em>bundle</em>.
</pre>

## Step 1. Installing Bumblebee

### Option A: Download runtime from Github repository<br>
Bumblebee is a single runtime.  You can get the latest, pre-built version for your platform in the “Releases” section at https://github.com/thoughtrealm/bumblebee.  Simply download and place the runtime in a common path in your OS.  You can place it in a directory and just execute it directly from there, but that can result in command lines that are longer than necessary.  It is recommended to place the runtime in a common path.

### Option B: Build and install using the Go compiler
If you have the Go compiler installed, you can clone the repo, then simply run “make install” in the root path of the repo.

If you are on Windows and do not have the make utility installed, you can run “go install” instead.  This build should work fine, with the one exception that the output of the “bee version” command will not be fully populated with build times.

### Validate Bumblebee is installed and working
Once installed, you can verify it is running correctly by simply typing...

    bee

That will output the root help info.  You can check the version by running...

    bee version

## Step 2. Initialize the Bumblebee Environment
The first step is to initialize the Bumblebee environment.  This will create the default profile,
populate the initial random key sets and some other artifacts that are required for sharing secrets.

To do so, just run...

    bee init

You will be asked about several options.

When asked, _“Enter a default sender key name or leave empty for none”_, provide a name you wish to
use for the default sender account in this profile.  It could be a name, a handle, an email address,
whatever you wish to use for identifying yourself.  The other user will be able to use whatever name
they wish to use in their user store for your identity.  Bumblebee will always validate the sending identity,
regardless of the name used to identify them in the user store.

However, in a formal environment, like in a corporate environment, it is recommended to use something unique
like your email address or an LDAP account name, etc.

Otherwise, for the other questions you are prompted for, just press enter for each to accept the default
options for now.

Once the initialization is completed, you can view the default profile identities by running...

    bee list keypairs

That will show you the public keys only for the default and system key pairs.

You can use the “**--show-all**” flag to see the seed and private keys as well…

    bee list keypairs --show-all

Of course, be aware that you must never share your private keys with anyone.  By default, they are
not printed out when listing the key pairs unless you provide the “**--show-all**” flag.

_**Note**: Bumblebee makes use of curve25519 key pair cryptography.
Specifically, it uses the **NKEYS** repo/packages (https://github.com/nats-io/nkeys).
**NKEYS** is provided by the **NATS** messaging server (https://nats.io/)._

_**Note**: Each identity is configured with two key pairs: a Cipher and a Signing key pair.
The Cipher key pair is a curve25519 key pair construction and is used for the encrypting
and decrypting processes.  The Signing key pair is an ed25519 key pair and is used for
signing secrets sent by that identity, so that the receiving user can validate the sender’s
identity.  The curve25519 support is found in the **XKEYS** package of the **NKEYs** repo._

You can also see the users that you have set up by running...

    bee list users

Of course, at this point, you will find that your user list is empty.  You must add or import
users to your local profile(s).  We will do so late in this Quick Start Guide.

## Step 3. Export your keys to share them with another user
To share secrets with another user, you must provide them with your public keys.  This can be
done easily by exporting your keys.  There are several ways to do this, but we will just focus
on exporting them to a file.

To do so, run the following command.  For _<username>_, use the name you provided in **Step 2**,
when you initialized the environment.  And when prompted for a password, just press return to
not provide a password for now.

    bee export user <username> --from-keypair --output-file export-user.txt

The export will have generated the file export-user.txt.  If you dump the file contents on Mac or Linux by using...

    cat export-user.txt

or on Windows by using...

    type export-user-txt

then, you will see something formatted similarly to the following, though with different values.

    :start  :export-user  :hex =====================================
    0086a44e616d65af7573657240646f6d61696e2e636f6da84461746154797065
    01aa43697068657253656564c0ab5369676e696e6753656564c0ac4369706865
    725075624b6579d938584349354e4f5a4649474c5a58334f4156425355515850
    32594d324445564d57574c474c4c4654524553594758575058414b3758425457
    4dad5369676e696e675075624b6579d93855443350573535525a585341355150
    344533424d535355414f594e494e4d564c4c414b4d5a525a4a36465950565657
    575144374144435344
    :end ===========================================================

Bumblebee uses that format because it is text safe.  Meaning, you can copy it and paste it into a 
message post, a slack post, an email body, etc.  It is simply the hex encoded sequence of the file’s
binary contents.  Bumblebee uses this encoding format for several different features.

If you wish, when exporting the file you may provide a password.  If you do so, the contents will be
symmetrically encrypted with XChacha20-poly1305 using Argon2 for key derivation.  This is a strong
cryptographic technique.

While the export data only contains public keys, you may wish to protect those by encrypting them
with a password.  If you do so, keep in mind that you **must** also provide the password to the user that
is importing your export file.  They will need the password in order to open and import the info.

Alternately, if you run the following command, it will output the data to the console instead
of a file.  This time, enter any password you wish when prompted for one.

    bee export user <username> --from-keypair --output-target console

With the output in the console, you can copy it and paste it to the other user.  Perhaps paste it in
their Slack channel.  And in this case, provide them with the password in some way.
If you provide it through some public transport outside of your trusted network, you will
probably want to use use a password to protect your public keys.  However, this will depend on your 
specific use case.

## Step 4. Import the other user’s keys in order to add them to your user store
After you supply your exported keys to another user, you will want to import their keys, so you can send them secrets.

To demonstrate this process, we will import your own keys as a user.

To do so, run the following command from the same directory as your export file.  Because the export
file was not exported with a password, you will not be prompted to enter a password.

When it asks you what name you want to use to import the user info, just press return to use the
exported name.  This will be your own name.  However, because your local entity's identity is stored
as a key pair set in the key pair store, it will not be a conflict to have a user with the same name
in the user store.

	bee import --input-file export-user.txt

Now, list your users again.

	bee list users

You should now see your name in the list of users.  Now, you can share files and secrets with that user.

## Step 5. Bundle a file for the other user
To keep this first step simple, copy any file you wish into the same directory.  We will refer to
the file using a name of “testfile.txt”.  You may change the name of your file so that the
commands are exactly as provided below, or you may leave your file named whatever it is and just
substitute the correct name in the commands you will be entering.

To encrypt a file for another user, we use the **bundle** command.  We supply the “**--to**” flag
to tell Bumblebee who the receiving user is so that it knows which keys to use for encrypting the
**bundle** header.

	bee bundle --input-file testfile.txt --to <username>

_**Note**: If you omit the “**--from**” flag, Bumblebee will use the default key pair identity as the
sender.  It is possible to have multiple local identities with Bumblebee, as well as multiple profiles,
which provide a separate security contexts. However, we will not go into that functionality here.
Just know that by omitting the “**--from**” flag, Bumblebee is signing the bundle using the key pair
named “default”.  You can refer to other docs for further info on multiple identities in the profile,
as well as multiple profiles._

That will have bundled **testfile.txt** into a new file, **testfile.bcomb**.

_**Note**: The ***.bcomb** extension refers to the “_Bumblebee combined_” bundle format.  We will not
go into the concept of _bundle types_ for now.  Just know that the _combined_ format means that the
two parts of a bundle, the header and the payload, are contained in the same file (or stream).
You can refer to other Bumblebee docs for an explanation of bundle formats._

Now, you can provide the bundled file to the other user.  Keep in mind that only the user specified with
the “**--to**” flag can decrypt the bundle, since they have corresponding private key to the public key
in your user store.  Not even you can decrypt the new bundle.

Of course, in this case, you are the other user.  Otherwise, you could send this file using whatever mechanism you wish.  You could attach it to an email, Slack it to them, etc.

## Step 6. Decrypt a bundled file from another user
For now, we will just decrypt the bundle using the same username, but the process is identical.

We can decrypt the bundle to a few different target outputs.  For now, we will decrypt the output to a
file.  If the input source of the bundle was a file, then Bumblebee will include the original file name
in the bundle header.  When decrypting that bundle to a file, Bumblebee will name the new, decrypted
file with the same name as the original file.

To demonstrate this, rename the current **testfile.txt** to something like **testfile.original.txt**.

Now, we use the ***open*** command to decrypt the bundle.

	bee open --input-file testfile.bcomb –from <username>

_**Note**: Similar to the bundle command, in this case, if you omit the “**--to**” flag, Bumblebee
will assume that the default key pair identity should be used as the receiving key pair. The receiver's private
key is used to decrypt the bundle.  Again, it is possible to have multiple local identities with Bumblebee,
as well as multiple profiles, which are basically separate security contexts; but, we will not go into
that functionality here.  Just know that by omitting the “**--to**” flag, Bumblebee is decrypting the bundle
using the local key pair named “***default***”.  You can refer to other docs for further info on multiple
identities in the profile, as well as multiple profiles._

The **--from** flag tells Bumblebee what signing keys to use to validate the sender’s identity.  When you import the other user’s public keys, you import their signing public key as well.  When opening a bundle, their public signing key is used to validate that they are the one who signed the bundle internally.  Bee does the for you and is why you must supply a --from reference.  If the user’s public key referenced by the “--from” flag does not validate correctly with the internally signed structures, then Bumblebee will output an error and will abort decrypting the bundle.

You should now see a new file with the same name as the original file, **testfile.txt**.  You can compare this
file to the original file that is now named **testfile.original.txt** using whatever process or comparison
command or tool that you want to use.  The two files should be identical.

## Share secrets that are not files
It is possible to send secrets to users that are not stored in a file.  There are a few ways to do this.
We will focus here on directly entering secrets from the console.  You can consult other docs for sharing
secrets in other ways.

For this _bundle_ we are going to enter the secret by typing it in.  We are also going to write it out to
the console, where we will copy and paste it to share with the other user.  They will use the pasted text to
decrypt your secret and write it to their console.

In so doing, we will share a secret that was never _explicitly_ written to a file.  

_**Note**: By "_explicitly_", we are acknowledging that all operating systems may use files for temporary
memory storage at any time.  So, you may be inadvertently writing data to the file system, even when you
don’t mean to do so._

To bundle a secret directly from the console, use the **--input-source console** flag value.

	bee bundle --input-source console --to <username>

This will provide a prompt to enter text.  You enter text, line by line.  You complete entering text by
just pressing **return** for an empty line.

Once you complete entering the text, <em>Bumblebee</em> will use that as the <em>bundle</em> input.

Also, since we did not provide any output flags, Bumblebee will default the output to the console as well.
Similar to the **export** output, you should see an output like this, though with different values than
this specific example.

Here's an example of the output...

<pre>
~/bee-demo : bee bundle --input-source console --to Bob
Enter one or more lines for the bundle input. Enter an empty line to stop. CTRL-C to cancel.
 : User root
 : password foo
 : 

Starting BUNDLE request...
:start :header+data ============================================
01ab786b76310bf0253cd66a8a24e2429ebc8ca5427708af6128c2b248209f80
234d66d56b4df5bdb57a0266b81233d32716c1ce6c71464a1c3724d7b3280109
1e396ea2e90b30968f738bccade31bc4257a9a93e4b65ac406e2ce72121ff48a
5293be085b2f74fa9f3015706f5f2a3a628119284a246305838b924f27f92b2e
27fbbc11fdd3f76db2d8504758be5c5a260493afa7a4d171c26e1054f0c2e575
c9631c2bb1ada972f9df75f3bdfd6a3f7bb9cf731906f2e97d86226fe0b39c5e
5510e1dc2beb18fd82ab419fa9b054af5654abe615f9a0e264e45278834bfdea
039bc655fe7e068191162b50db4feb174bd24b7826af1ef8c2bc2f5e5b0c73e4
57f397dbe53b7e0f61ced9e3ea0073f7a7c106629fcb3a3fc946ae7cc247dc76
f6b5e2d091863012303601ac713af6b548e96e95c9d854a20f1321bf458f67e5
616ba98f9f24a8a5dc89d7ab57354abbcbbfe347e19f776850db1f83f1df98fe
76c2b9a4acce7266164a8d5436f2027b6c96f70e8f2b2a5b568ecf7aea688b5c
71396c4af0fd8dd2694e043ed9ead58e47e9d7e255c2c2c2139a5575ac8e416c
23b6ced87f95092ed77641edd9895805c4910a2d51469888fc9d15b641d8bf42
fc6f423a4cbd62c0e1fd1515e2e8650bd479deab02405a1482503bbf99011cd6
e9a1810979d56b1042c59b
:end ===========================================================
BUNDLE completed. Bytes written: 491 in 83 milliseconds.
~/bee-demo : 
</pre>

That is a text safe version of the bundle data.  Meaning, it can be safely entered into any app or service
that just supports text.  For example, you could paste it into a text, email or Slack post.

For this example, let's just copy the output to the clipboard.  Be sure to include the lines beginning 
with the line that starts with "**:start**" and down to the line that starts with "**:end**".

Now, there are several ways to **open** the clipboard bundle depending on the operating system.

For convenience on Windows or Mac ARM64, Bumblebee provides an input-source of **clipboard**, 
like in this example...

    open --input-source clipboard --output-target console --from <username>

On Debian with xclip installed, you can use...

    xclip -o | bee open --output-target console --from <username>

Otherwise, for any other operating system, you would use whatever command formation is needed, such 
as **pbpaste** on Mac.  Linux has a number of options for clipboard access, so refer to whatever your
distro or environment needs.

Regardless, you should see something like the following output...

<pre>
~/bee-demo : xclip -o | bee open --output-target console --from Bob
Starting OPEN request...

Decoded data...
==========================================================
User root
password foo
==========================================================

OPEN completed. Bytes written: 22 in 61 milliseconds.
~/bee-demo : 
</pre>

Generally, copying and pasting like this would only be appropriate for smaller bundles, such as in the example
when sharing credentials or something like that.  While you could use clipboard sharing for much larger
bundles, the approach described in the test above would not be practical.

The point of that example is simply to show sharing a bundle without writing it to a file explicitly.

_**Note**: For builds targeting Windows and Mac ARM64, Bumblebee provides a target of clipboard using 
**--output-target clipboard**.  That will write the output directly to the clipboard._

## Storing secrets locally using **--local-keys**
Bumblebee provides  