# OneWipe Core - Secure Data Destruction Tool

## Quick Start

**Build:**
```bash
gcc -o onewipe_core onewipe_core.c -lcrypto
```

**Usage:**
```bash
# List devices
sudo ./onewipe_core list

# Wipe device (DESTRUCTIVE!)
sudo ./onewipe_core overwrite /dev/sdX 1 255

# Random wipe
sudo ./onewipe_core overwrite /dev/sdX 1 rand

# Generate certificate
./onewipe_core gen-cert logfile.log cert.json

# Sign certificate
./onewipe_core sign-cert cert.json private.pem cert.sig

# Verify signature
./onewipe_core verify-cert cert.json cert.sig public.pem
```

## Requirements
- Linux with root access
- OpenSSL development libraries (`libssl-dev`)

## Warning
⚠️ **DESTRUCTIVE TOOL** - Will permanently erase data. Test only on disposable devices.

## Output
- Logs: `./onewipe-logs/`
- Certificates: `./onewipe-