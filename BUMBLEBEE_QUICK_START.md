# _Bumblebee_ Quick Start Guide
#### _Last updated: Dec 10, 2023_

This document will provide a set of steps that will guide you through installing _Bumblebee_,
sharing keys with another user, and then demonstrates a few methods for sharing secrets.

This guide is focused on simple use cases that will show you basic functionality in _Bumblebee_.
While there are a number of additional features and options in _Bumblebee_, this document will not go
into those details to a great degree.

The examples shown are taken from sessions on _Debian ARM64_.  The _Bumblebee_ syntax itself is identical
across all platforms of _Windows_, _Linux_, and _Mac_.  However, the command line constructions may
differ a bit due to OS distinctions, such as examples that pipe clipboard input. Also, I will use the
term _clipboard_ to refer to the corresponding functionality for any operating system, regardless of what
it is called. 

## Sending a file to another user
We will go through a simple scenario of sending a file to another user.  While there are ways to share
secrets without using files, we will focus on sharing files first.

Before we can do so, we will need to install _Bumblebee_, as well as set up a user that
we wish to share secrets with.

After the setup is done, we will walk through steps to demonstrate the general pattern for
sharing files with other _Bumblebee_ users.  

The basic pattern is as follows:
<pre>
1. Run the <b>bundle</b> command and create the encrypted <em>bundle</em> for another user.  
2. Supply the encrypted <em>bundle</em> to that user.
3. The other user then decrypts the <em>bundle</em> with their private keys using the <b>open</b> command. 
4. <em>Bumblebee</em> validates the sender's identity when it opens the <em>bundle</em>.
</pre>

_**Note**: In this guide, we will be setting up a scenario where the sender and receiver are the same identity.  This
is so that you can do these steps without requiring another user.  However, the steps themselves are the same
when interacting with other users._

## Step 1. Installing _Bumblebee_

### Option A: Download binary or archive from GitHub repository
_Bumblebee_ is a single runtime.  You can download the latest, pre-built binary or archive for your platform
in the “Releases” section of the GitHub project [here](https://github.com/thoughtrealm/bumblebee/releases/latest).

_**NOTE**: Be sure to validate the binary and/or archive with the hashes provided.  They can be found in the
release page description, as well as attached as a project artifact, or in the related description file
in the project's build/ path._

Simply download and place the binary in a common path in your OS.  You can place it in a directory and
execute it directly from there, but that can result in command line constructions that are longer than
necessary, depending on your OS. Therefore, it is recommended to place the binary in a common path.

### Option B: Build and install using the Go compiler
If you have the Go compiler installed, you can clone the repo, then run “make install” in the
root path of the repo.

If you are on _Windows_ and do not have the make utility installed, you can run ***go install*** instead.
That form of build should work fine, with the one exception that the output of the ***bee version***
command will not be populated with build times.

### Validate _Bumblebee_ is installed and working
Once installed, you can verify it is running correctly by typing...

    bee

That will output the root help info.  You can check the version by running...

    bee version

## Step 2. Initialize the _Bumblebee_ Environment
The first step is to initialize the _Bumblebee_ environment.  This process will do the following...
1. Create the default profile
2. Create the profile's **User Store** and **Identity Store** 
3. Create a default identity and related key sets
4. Create additional elements in the profile that are required for sharing secrets.

To do so, just run...

    bee init

You will be asked about a few options.

When asked, _“Enter a default sender key name or leave empty for none”_, provide the name you wish to
use for the default sender account in this profile.  It could be a name, a handle, an email address,
whatever you wish to use for identifying yourself.  Keep in mind that other users will be able to use
whatever name they wish to use in their user stores for your identity.  _BumbleBee_ will always validate
the sending identity by key sets, regardless of the name used to identify them in the user store.

If setting up a formal environment, such as in a corporate setting, it is recommended to use something unique
for your name, such as your email address or an LDAP account name, etc.

For any other questions, just press _return_ for each to accept the default options.

Once the initialization is completed, you can view the default profile identities by running...

    bee list keypairs

That will show you the public keys only for the default and system identities.

You can use the ***--show-all*** flag to see the seed and private keys as well...

    bee list keypairs --show-all

Be aware that you must **NEVER** share your private keys with anyone.  By default, they are
not printed out when listing the key pairs unless you provide the ***--show-all*** flag.

_**Note**: _BumbleBee_ makes use of curve25519 key pair cryptography.
Specifically, it uses the **NKEYS** packages (https://github.com/nats-io/nkeys).
**NKEYS** is provided by the **NATS** messaging server (https://nats.io/)._

_**Note**: Each identity is configured with two key pairs: a Cipher and a Signing key pair.
The Cipher key pair is a curve25519 key pair construction and is used for encrypting
and decrypting processes.  The Signing key pair is an ed25519 key pair and is used for
signing secrets sent by that identity.  This allows the receiving user to validate the sender’s
identity.  The curve25519 support is found in the **XKEYS** package of the **NKEYs** repo._

You can see the users that have been set up by running...

    bee list users

Of course, at this point, you will find that your user list is empty.  You must add or import
users to your local profile(s).

## Step 3. Export your public keys to share them with another user
To share secrets with another user, you must provide them with your public keys.  This can be
done easily by exporting your keys.  There are several ways to do this, but we will just focus
on exporting them to a file.

To do so, run the following command.  For the argument _<username>_, use the name you provided
in **Step 2**, when you initialized the environment.  And when prompted for a password, just
press _return_ to not provide a password.

    bee export user <username> --from-keypair --output-file export-user.txt

The export will have generated the file export-user.txt.  If you dump the file contents on _Mac_ or _Linux_
by using...

    cat export-user.txt

or on _Windows_ by using...

    type export-user.txt

You will see output formatted similarly to the following, though with different hex values.

    :start  :export-user  :hex =====================================
    0086a44e616d65af7573657240646f6d61696e2e636f6da84461746154797065
    01aa43697068657253656564c0ab5369676e696e6753656564c0ac4369706865
    725075624b6579d938584349354e4f5a4649474c5a58334f4156425355515850
    32594d324445564d57574c474c4c4654524553594758575058414b3758425457
    4dad5369676e696e675075624b6579d93855443350573535525a585341355150
    344533424d535355414f594e494e4d564c4c414b4d5a525a4a36465950565657
    575144374144435344
    :end ===========================================================

_BumbleBee_ uses that format because it is text safe.  Meaning, you can copy it and paste it into a 
message post, a Slack post, an email body, etc.  It is simply the hex encoded sequence of the file’s
binary contents.  _BumbleBee_ uses this encoding format for several features.

If you wish, when exporting your keys, you may provide a password for the file.  If you do so,
the contents will be symmetrically encrypted with XChacha20-poly1305 using Argon2 for key derivation.
This is a strong cryptographic technique.

While a password is not required, since the export data only contains public keys, you may wish to
protect those by encrypting them with a password.  If you do so, keep in mind that you **must** also
provide the password to the user that is importing your export file.  They will need the password to
open and import the info.

Another option is to run the following command, which will output the data to the console instead
of a file.

    bee export user <username> --from-keypair --output-target console

With the output in the console, you can copy it and paste it to the other user.  Perhaps paste it in
their Slack channel.  If you provide it through some public transport outside your trusted network, you will
probably want to use a password to protect your public keys.  This will depend on your specific use case.

## Step 4. Import the other user’s public keys in order to add them to your user store
After you supply your exported keys to another user, you will want to import their keys, so you can
send them secrets.

To demonstrate this process, we will import your own keys as a user.  However, the process is the same
for other users.

To do so, run the following command from the same directory as your export file.  Because the export
file was not exported with a password, you will not be prompted to enter a password.  If you did 
provide a password, you will be prompted to enter it.

When it asks you what name you want to use to import the user info, just press _return_ to use the
exported name.  This will be your own name.  However, this is not an issue, because your local
identity is stored as a key pair set in the key pair store.  This is not a conflict with the same name
in the user store.

	bee import --input-file export-user.txt

Now, list your users again.

	bee list users

You should see your name in the list of users.  You can now share secrets with that user.

## Step 5. Bundle a file for the other user
To keep this first step simple, copy any file you wish into the same directory.  We will refer to
that file using a name of “testfile.txt”.  You may change the name of your file to this,
so that the commands exactly match what is as provided below.  Or, you may leave your file named whatever
it is and substitute the correct name in the commands you will be entering.

To encrypt a file for another user, we use the **bundle** command.  We supply the “**--to**” flag
to tell _BumbleBee_ who the receiving user is so that it knows which keys to use for encrypting the
**bundle** header.

	bee bundle --input-file testfile.txt --to <username>

_**Note**: When you omit the “**--from**” flag, _BumbleBee_ will use the default key pair identity as the
sender.  It is possible to have multiple local identities with _BumbleBee_, as well as multiple profiles,
which provide a separate security context. However, we will not go into that functionality yet.
Just know that by omitting the “**--from**” flag, _BumbleBee_ is signing the bundle using the key pair
named “default”.  We will demonstrate multiple identities and profiles later in this guide._

That will have bundled **testfile.txt** into a new file, **testfile.bcomb**.

_**Note**: The ***.bcomb** extension refers to the “__BumbleBee_ combined_” bundle format.  We will not
go into the details of _bundle types_ for now.  Just know that the _combined_ format means that the
two parts of a bundle, the header and the payload, are contained in the same file (or stream).
You can refer to other _BumbleBee_ docs for an explanation of bundle formats._

Now, you can provide the bundled file to the other user.  Keep in mind that only the user specified with
the “**--to**” flag can decrypt the bundle, since they alone have the corresponding private key relating
to the public key in your user store.  Not even you can decrypt the new bundle.

In this case, you are the other user.  Otherwise, you would send this file using whatever mechanism you
wish.  You could attach it to an email, Slack it to them, etc.

## Step 6. Decrypt a bundled file from another user
For now, we will just decrypt the bundle using your user info, but the process is identical.

We can decrypt the bundle to a few different output targets.  For now, we will decrypt the output to a
file.  

If the input source of the bundle is a file, then _BumbleBee_ will include the original file name
in the bundle header.  Then, when the other user decrypts the bundle to a file, _BumbleBee_ will attempt
to name the new, decrypted file with the same name as the original file.

To demonstrate this, rename the current **testfile.txt** to something like **testfile.original.txt**.

Now, use the ***open*** command to decrypt the bundle.

	bee open --input-file testfile.bcomb –-from <username>

_**Note**: Similar to the bundle command, in this case, if you omit the “**--to**” flag, _BumbleBee_
will assume that the default key pair identity should be used as the receiving key pair. The receiver's private
key is used to decrypt the bundle.  Again, it is possible to have multiple local identities with _BumbleBee_,
as well as multiple profiles, which are basically separate security contexts; but, we will not go into
that functionality yet.  Just know that by omitting the “**--to**” flag, _BumbleBee_ is decrypting the bundle
using the local key pair named “***default***”.  We will demonstrate multiple identities and profiles
later in this guide._

The **--from** flag tells _BumbleBee_ which signing keys to use to validate the sender’s identity.  When
you imported the other user’s public keys, you imported both their cipher and signing keys.  When
opening a bundle, their signing key is used to validate that they are indeed the one who sent the
bundle.  

_Bumblebee_ does this for you, so you must supply the **--from** reference.  If the _bundle_'s signature
does not match the **--from** user's identity, then _BumbleBee_ will output an error and will
abort decrypting the bundle.

You should now see a new file with the same name as the original file, **testfile.txt**.  You can compare this
file to the original file that is now named **testfile.original.txt** using whatever process, comparison
command, or tool that you wish.  The two files should be identical.

## Share secrets that are not files
It is possible to send secrets to users that are not stored in a file.  There are a few ways to do this.
We will focus here on directly entering secrets from the console.  You can consult other docs for sharing
secrets in other ways.

For this _bundle_ we are going to enter the secret by typing it in.  We are also going to write it out to
the console, where we will copy and paste it to share with the other user.  They will use the pasted text to
decrypt your secret and write it to their console.

In so doing, we will be sharing a secret which was never _explicitly_ written to a file.  

_**Note**: By "_explicitly_", we are acknowledging that all operating systems may use files for temporary
memory storage at any time.  So, you may be inadvertently writing data to the file system, even when you
don’t mean to do so._

To bundle a secret directly from the console, use the **--input-source console** flag value.

	bee bundle --input-source console --to <username>

Bumblebee will provide a prompt for entering text, which you enter one line at a time.  You complete 
entering text by just pressing **return** on an empty line.

Once you complete entering the text, _Bumblebee_ will use that as the input data for the _bundle_.

Also, since we did not provide the **--output-target** flag, _BumbleBee_ will default the output
to the console as well. 

Similar to the output you saw with the **export** comand, you should see an output like the following, 
though with different hex values than this specific example.

<pre>
~/bee-demo : <b>bee bundle --input-source console --to Bob</b>
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
that just supports text.  For example, you could paste it into a text message, text document, email
or Slack post.

For this example, let's copy the output to the clipboard.  Be sure to include the lines beginning 
with the line that starts with "**:start**" and down to the line that starts with "**:end**".

There are several ways to **open** the _bundle_'s data in the clipboard, depending on the operating system.

For convenience on _Windows_ or _Mac ARM64_ builds, _BumbleBee_ provides an input-source of **clipboard**, 
as shown in this example...

    open --input-source clipboard --output-target console --from <username>

However, on _Debian_ with xclip installed, you can use...

    xclip -o | bee open --output-target console --from <username>

Otherwise, for any other operating system or configuration, you would use whatever command formation is
needed, such as using **pbpaste** on _Mac_.  _Linux_ has a number of options for clipboard access,
so refer to whatever your distro or environment needs.

Regardless, you should see something like the following output...

<pre>
~/bee-demo : <b>xclip -o | bee open --output-target console --from Bob</b>
Starting OPEN request...

Decoded data...
==========================================================
User root
password foo
==========================================================

OPEN completed. Bytes written: 22 in 61 milliseconds.
~/bee-demo : 
</pre>

Generally, copying and pasting like this would only be appropriate for smaller bundles, such as in this
example of sharing credentials or something like that.  While you could use clipboard sharing for much
larger bundles, the approach described in the example above may not be practical.

The point of that example is simply to show how to share a bundle without writing it to a file explicitly.

_**Note**: For builds targeting **Windows** and **Mac ARM64**, _BumbleBee_ provides the flag value
**--output-target clipboard**.  That flag value tells _BumbleBee_ to write the output directly to the
clipboard.  When using that option, you do not have to copy the console output, since _Bumblebee writes it to
the clipboard for you._

## Storing secrets locally using **--local-keys**
Sometimes you wish to encrypt files for your own purposes.  Perhaps, you have important documents
that you want to keep in an encrypted state.  Maybe you are storing them in a Cloud storage service and
you want to encrypt them before doing so.  Or maybe you want to encrypt files for local backup purposes.

To support this, _BumbleBee_'s **bundle** and **open** commands support a flag **--local-keys**.  When you pass
this flag to the **bundle** and **open** commands, _BumbleBee_ will use the profile's _read_ and _write_
system key pairs.  

When you run the command...

<pre>
bee list keypairs
</pre>

You will see the system key pairs output with the names **keystore_read** and **keystore_write**.  These
are used for supporting the **--local-keys** functionality.  They are also used to encrypt your local
**User Store**.

To use these system keys, provide the **--local-keys** flag to any **bundle** or **open** command 
construction.

Similar to the prior example where we used the console for input, here's an example using **--local-keys** ...

<pre>
~/bee-demo : <b>bee bundle --input-source console --output-file test --local-keys</b>
Enter one or more lines for the bundle input. Enter an empty line to stop. CTRL-C to cancel.
 : User root
 : password fbar2
 : 

Starting BUNDLE request...
BUNDLE completed. Bytes written: 506 in 79 milliseconds.
~/bee-demo : <b>ls</b>
test.bcomb
~/bee-demo : 
</pre>

Notice the newly created bundle file **test.bcomb**.

To decode that bundle file, you would also use **--local-keys** like this...

<pre>
~/bee-demo : <b>bee open --input-file test.bcomb --output-target console --local-keys</b>
Starting OPEN request...

Decoded data...
==========================================================
User root
password fbar2
==========================================================

OPEN completed. Bytes written: 24 in 61 milliseconds.
~/bee-demo :
</pre>

Notice that there are no references to the flags **--to** or **--from**.  When you use the 
**--local-keys** flag, _BumbleBee_ is simply substituting the **--to** and **--from** key references with
the system keys accordingly, depending on the command being using.

The files you encrypt with the **--local-keys** option can be stored offsite or in your backups or
wherever you wish.  You just retrieve them as needed and open them accordingly.

**Keep in mind** that bundles built with the **--local-keys** flag **can only be decrypted** using the
**--local-keys** option in an environment with the specific system keys that were used to bundle the data.
So, for any data encrypted with the **--local-keys** flag, be sure to backup the profile or maintain the
profile's environment that you used to bundle that data.  If you lose the environment in some way and have
no backup, you **will not be able to decrypt** those files.  Of course, this is true of any bundled data,
regardless of the key pairs and identities used.

## Using multiple key pair identities in the same profile
It is possible to create any number of identities within a single profile.  For this explanation, we will
refer to the ***default*** profile.  However, this can be done with any profile.

Let's say you want to maintain additional identities, maybe one for work and one for home.  For this example,
we'll assume that these would be in addition to the default identity created when you initialized the 
***default*** profile.

To do so, you would add a new identity like this...

<pre>
bee add keypair home
</pre>

And...

<pre>
bee add keypair work
</pre>

_BumbleBee_ will create the new identities and output the key pair info.

Now, when you run the command...

<pre>
bee list keypairs
</pre>

You will see the new **home** and **work** key pairs.

You can export these identities to other users like this...

<pre>
bee export user home --from-keypair --output-file export-home.txt
</pre>

To reference this identity when creating a bundle, specify it with the **--from** flag...

<pre>
bee bundle --input-source console --output-file test --from home --to Bob
</pre>

The other user would import your identity like it would for any user, and then reference it when opening
the bundle, like this...

<pre>
bee open --input-file test.bcomb --output-target console --from home
</pre>

Of course, the other use may use a different name for your identity than just "**home**",
such as "BobHome" or something like that.

Using this process, you can add as many identities in a single profile as you wish.  You can
remove them using **bee remove keypair**.

## Using multiple profiles
_BumbleBee_ allows you to create any number of profiles.  Each profile provides its own security
context of user and key pair stores.  So, any particular profile only sees the users and key pair identities
that were set up in that profile.

For example, instead of creating new identities for **work** and **home** in the same profile as we did
in the prior example, we can create completely separate profiles for work and home.

To do so, create a new profile with the following command.  This is the same process as the
**init** command that we demonstrated earlier, so you can answer the questions as you would like
when creating the new profile's environment.

<pre>
bee add profile home
</pre>

Now, you can list your profiles using this command...

<pre>
bee list profiles
</pre>

You will see your **default** profile, as well as the new **home** profile.

You can use the profile in a couple of ways.

### Setting the active profile with the **use** command

One way is by setting the active profile using the **--use** flag as follows...

<pre>
bee use home
</pre>

You can then use this command to see the currently active profile...

<pre>
bee show profile
</pre>

When you run a _BumbleBee_ command, it uses the active profile.  So, if you change the active
profile to **home** using the ***bee use <profile>*** command, then any commands you run will
do so in the **home** profile.

For example, any users or identities you add will be added to that profile.  If you then switch to another
profile, those users and identities will not be there.

To change the active profile back to the _default_ profile, simply run ***bee use default***.

### Referencing a different profile with the **--use** flag
Another way to reference a different profile is to provide the ***--use <profile>*** flag with any command.

This flag allows you to run any command in the context of another profile, without changing the active 
profile.

For example...

    bee bundle --input-source console --output-file test --to <username> --use home

That command will bundle the file using the default identity in the **home** profile, as well as the user
info in the **home** profile's user store.

You can also see the profile config for any profile like this...

    bee show profile [profilename]

Notice that command does not require the **--use** flag.  If **profilename** is provided, then 
that profile is shown.  If not provided, then it will show the active profile info.

## Wrapping up
That wraps up this Quick Start Guide.  For more detailed info, see the [_Bumblebee User Guide_](USER_GUIDE.md)
or other docs as needed.