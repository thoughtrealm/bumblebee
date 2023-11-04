### !IMPORTANT NOTE!
_Please be aware that this project contains strong cryptographic capabilities.
You must determine if the cryptographic functionality in this product is legal in your 
municipality.  If not, you are responsible for not downloading, building or using
this product._  

_The **Bumblebee Authors** assume no liability or responsibility, either legally or
ethically, for any particular usage of this product.  This product is provided for
ethical application and usage only._

## Bumblebee Overview
**Bumblebee** is a system for sharing secrets, which provides functionality for easily encrypting 
files, as well as text or binary inputs. These can then be provided to other **Bumblebee** users 
in various forms for decrypting.

**Bumblebee** runs in common desktop environments, including Mac, Linux and Windows.
Future implementations may be added for mobile app support.

**Bumblebee** utilizes well-known asymmetric and symmetric technologies, which will be described
in a later section and documents. This approach allows you to easily encrypt small and large files 
for sharing with other users.

Installing **Bumblebee** sets up a local environment and initializes various cipher components,
including a default profile, several requisite key sets, and corresponding key stores.

Similar to other app environments, you can provide your public key(s) to other users for adding
to their key store, as well as add their public keys to your key store. This allows you to securely
share secrets with each other.

The local environment setup may be sufficient for all of your simple use cases; however, a 
public key management service will be provided in the future, so you don't have to add keys manually, 
as well as to provide for more complex use cases.

If desired, you can easily create multiple named key sets for securing assets with specific keys
for specific users or groups of users.

Also, **Bumblebee** will allow you to create a number of local profiles, which will allow you to 
isolate multiple secret sharing domains. Each profile maintains a distinct set of key stores and
user entities.

**Bumblebee** supports emitting output in a text safe form to the console or clipboard, which can 
be copied and pasted into services like Slack, allowing the other user to simply copy the 
encrypted text and decrypt it with **Bumblebee**.  You could also paste it into an email body, 
or attach a secret saved as an encrypted file to an email, etc.

This local mode can operate in a disconnected, offline state.  This is so that you can encrypt files 
offline for other users, then provide them using any simple transport mechanism, like 
USB drives.  Of course, once you are back online, you can deliver them digitally.

It is intended to provide a server in the future that will support distributing your secrets, 
which will make sharing secrets with other users easier.

## Status
This system is still in initial development. The primary use cases are completed, and
testing is provided for critical paths, including all cipher constructs. Primary functional paths
for providing and processing secrets are complete.  What is missing is mostly related to
in-code and external documentation, as well as additional testing for non-critical paths.

The planned server functionality described above will be completed in the next phase, 
once the initial "local" or "stand alone" functionality is sufficiently complete.

The product is in a state that should be safe to use for secret sharing, though all 
desired documentation and non-critical functionality may not be available at this time. 

For specific details on the current status, see [Status](STATUS.md).

For a general roadmap of future functionality plans, see [Road Map](ROAD_MAP.md)

Given this status, there may be bugs and missing or incomplete features. Sufficient testing has been
completed, such that any remaining issues should be in non-critical areas.  Both manual analysis
and unit testing efforts are employed to validate correct performance.

Of course, if find or experience any issues, please report them as described below in 
[Reporting bugs and making feature requests](#Reporting_bugs_and_making_feature_requests).

## Design, Security Approach and Technical Details
Generally, the design for this system creates no proprietary or unique cryptographic approaches.
All crypto-related functionality simply uses well-known industry algorithms and approaches.
**Bumblebee** simply wraps these known algorithms and approaches with an easy-to-use environment. 

I've incorporated the crypto functionality using patterns and approaches I've personally used
for decades.  I do not believe any significant risk is exposed, given a modicum of 
attention to local security.

Having said that, you are welcome to analyze my design and code, as well as the crypto and 
supporting functionality.  If you uncover any concerns, feel free to report them as described 
below. I will address any practically valid concerns that are reported. 

You will find details of crypto and related tech [here](CRYPTO_DETAILS.md).

You will find details of file construction and formats [here](FILE_DETAILS.md).

You will find a threat analysis/model [here](THREAT_DETAILS.md)

## Installing and setting up the **Bee** runtime
Bumble functionality for supporting stand alone, local behavior is contained in the 
**bee** runtime.  No other dependencies are required.

For runtime functionality, you can simply place the **bee** runtime in any path you wish.

You can obtain the **bee** runtime from the github repo, or you can build and/or install it 
yourself if you have the GO environment installed.

### Downloading from github

### Remote build

### Local build




## How to use
For simple steps to get up and running, see the [Quick Start](QUICK_START.md).

For detailed docs, see the [User Guide](USER_GUIDE.md).

In addition, I have prepared a demo video of how to use Bumblebee.  It is hosted on YouTube
[here](youtube_url).

## License
**Bumblebee** uses an MIT license.  For license details see [License](LICENSE).

## Reporting bugs and making feature requests
Please create an Issue for any bugs you find or suggestions you may have relating to
**Bumblebee** functionality. I will try to respond to these as quickly as I can.

When creating issues for bugs, please prefix the title with "Bug:", like "Bug: Blah Blah feature is not working right."

And for feature requests, please prefix the title with "Feature Request:", like "Feature Request: Adding blah blah functionality would make this utility such the major hotness"

## Contributing
If you wish to contribute, you may fork and submit pull requests. 
Please follow this GitHub guide to do so: 
[GitHub: Contributing to Projects](https://docs.github.com/en/get-started/quickstart/contributing-to-projects) 

I will review those as I have time.