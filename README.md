### **!IMPORTANT NOTE!**
_Please be aware that this project contains **strong cryptographic capabilities**.
You must determine if the cryptographic functionality in this product is legal in your 
country/state/city/municipality/whatever.  If not, **you are responsible** to not download, 
build or use this product._  

_The **Bumblebee Authors** will assume no liability or responsibility, either legally or
ethically, for any particular usage of this product.  **This product is provided for
ethical application and usage only**._

## What Is Bumblebee
_Bumblebee_...<br>

- Allows you to encrypt files and messages to share with specific Bumblebee users.

- Utilizes **hybrid cryptography**, combining both **asymmetric** and **symmetric**
cryptographic functionality.

- Uses curve25519 and ed25519 for asymmetric functionality. Uses XChacha20-Poly1305 and Argon2
for symmetric needs. Random sequences are generated by Go's crypto/rand package.

- Uses signing functionality so that the receiver is able to validate the sender's identity.

- Is a single binary that runs on _Mac_, _Windows_ and _Linux_.  Mobile support may be provided in
the future.

- Can encrypt large files, small files, keyboard text entry, clipboard inputs, and pipe inputs.

- Can render encrypted text and binary secrets into text safe formats that can be safely embedded into 
text docs, text messages, Slack posts, email, etc., which can be copied and decrypted by the intended user,
without requiring a file exchange.

## Bumblebee Guides and Docs

- Currently, the [Quick Start Guide](BUMBLEBEE_QUICK_START.md) is the best doc to get you up to speed
on Bumblebee functionality. It will take you through a brief description of the current
installation options, show you how to set up and initialize Bumblebee, plus walk you through some
examples of sharing secret files and messages.

- The [User Guide](USER_GUIDE.md) is not available yet.  For now, please refer to
[Quick Start Guide](BUMBLEBEE_QUICK_START.md).

- The latest binaries and archives are [here](https://github.com/thoughtrealm/bumblebee/releases/latest).
Refer to the [Quick Start Guide](BUMBLEBEE_QUICK_START.md) for how to install.

- For information on the current _Bumblebee_ project status, see [this](STATUS.md).

- Bumblebee is a CLI.  For the status of the current command implementations, see the
[Command Implementation Status](COMMAND_DEFINITIONS.md).

- For info on the future goals of this project, see the [Project Road Map](ROAD_MAP.md).

- The [Technical Details](TECHNICAL_DETAILS.md) document provides a detailed description of the 
crypto implementation logic, the specific crypto algorithms used, and their corresponding code sources.

- There is a brief [YouTube Overview Presentation](https://youtu.be/d8Ay6VDspAk) that discusses the general goals and status of
the _Bumblebee_ project.

- There is also a [YouTube Basic Demo](https://youtu.be/9ceIijof4eI) that demonstrates some of the common usage scenarios of _Bumblebee_.

## Bumblebee Project Status
- The initial CLI environment is still in progress.
- MVP flows are complete and tested, including critical cipher paths.
- It is ok to use now.
- Nevertheless, you may run into bugs. If so, please report them as described [here](#reporting).

## An Overview Of The Bumblebee Project
_BumbleBee_ is a system for sharing secrets.
It provides functionality for encrypting and sharing files, as well as text or binary inputs. 
These can then be provided to the _BumbleBee_ user for decrypting.  These can be provided in
various ways and in various forms.

_BumbleBee_ runs in common desktop environments, including _Mac_, _Linux_ and _Windows_.
Future support for mobile environments may be provided.

_BumbleBee_ is a _Hybrid Encryption_ tool, utilizing well-known asymmetric and symmetric technologies. 
This approach allows you to encrypt small and large files for sharing with other users.

Installing _BumbleBee_ sets up a local environment and initializes various cipher components,
including a default profile, several required key sets and identities, and corresponding key stores.

Similar to some other tools, you will share public keys with other _Bumblebee_ users.
This allows you to securely share secrets with each other.

_BumbleBee_ supports emitting output in a text safe form to the console or clipboard, which can
be copied and pasted into services like Slack, or text messages, etc.
This allows the other user to simply copy the encrypted text and decrypt it with _BumbleBee_.
You could also paste it into an email body, or attach a secret saved as an encrypted file to an email, etc.

_Bumblebee_'s local environment does not currently require any online access.
This has the benefit of allowing you to encrypt files and secrets for other users, while you are
offline.  Secrets can be provided using any transport mechanism, like USB drives.
Of course, when you are online, you can deliver them digitally.

The basic functionality provided by the _Bumblebee_ CLI may be sufficient for all of your use cases.
However, there are some more advanced features if you should need them.

If desired, you can add multiple key set identities for securing assets with specific keys
for specific users, or groups of users.

Also, _BumbleBee_ allows you to create any number of profiles, which will allow you to 
isolate multiple secret sharing or security domains. Each profile maintains a distinct set of identities and
user references.  

Also, a key management service may be provided in the future, so that you don't have to import
keys manually.  That service may also support more complex use cases and transport options.

## Status
This system is still in initial development. The primary use cases are completed, and
testing is provided for critical paths, including all cipher constructs. Primary functional paths
for providing and processing secrets are complete.  What is missing is generally related to
in-code and external documentation, as well as additional testing for non-critical paths.

The planned service functionality mentioned previously will be completed in the next phase, 
once the initial "local" or "stand alone" functionality is sufficiently completed.

The project is in a state that should be safe to use for secret sharing, though all 
desired documentation and non-critical functionality may not be available at this time. 

For specific details on the current status, see [Status](STATUS.md).

For a roadmap of future project plans, see the [Road Map](ROAD_MAP.md)

Given the current status, you may encounter bugs and missing or incomplete features.
Sufficient testing has been completed, such that any remaining issues should be in non-critical areas.
Both manual analysis and unit tests are used to validate functionality.

If you find or experience any issues, please report them as described below in 
[Reporting bugs and making feature requests](#reporting).

## Design, Security Approach and Technical Details
The design for this system creates no proprietary or unique cryptographic functionality.
All crypto-related functionality uses well-known industry algorithms and approaches.
_BumbleBee_ simply wraps these known algorithms and approaches with an easy-to-use environment. 

Having said that, you are welcome to analyze the design and code, as well as the crypto and 
supporting functionality.  If you uncover any concerns, feel free to report them as described 
below.  Any practically valid concerns that are reported will be addressed. 

You will find details of crypto and related tech in the [Technical Details](TECHNICAL_DETAILS.md) document.

You will find a some details of file construction and formats [here](docs/StreamCompositionOfBundles.pdf).

A Threat Model document is not available yet.  We do intend to provide one.

## License
_BumbleBee_ is licensed under an MIT license.  For license details see [License](LICENSE).

## <a name="reporting"></a>Reporting bugs and making feature requests
Please create an Issue for any bugs you find or suggestions you may have relating to
_BumbleBee_ functionality. I will try to respond to these as quickly as I can.

When creating issues for bugs, please prefix the title with "Bug:", like "Bug: Blah Blah feature is not working right."

And for feature requests, please prefix the title with "Feature Request:", like "Feature Request: Adding blah blah functionality would make this utility such the major hotness"

## Contributing
If you wish to contribute, you may fork and submit pull requests. 
Please follow this GitHub guide to do so: 
[GitHub: Contributing to Projects](https://docs.github.com/en/get-started/quickstart/contributing-to-projects) 

Those will be reviewed as time permits.