import subprocess
import sys
import os

def run_cmd(cmd):
    try:
        result = subprocess.run(cmd, shell=True, check=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True)
        return result.stdout.strip()
    except subprocess.CalledProcessError as e:
        print(f"[ERROR] Command failed: {cmd}")
        print(e.stderr.strip())
        return None

def list_sata_devices():
    output = run_cmd("lsblk -d -o NAME,TYPE,TRAN")
    sata_devices = []
    if output:
        lines = output.splitlines()
        for line in lines[1:]:  # Skip header line
            parts = line.split()
            if len(parts) == 3:
                name, type_, tran = parts
                if type_ == "disk" and tran == "sata":
                    sata_devices.append(f"/dev/{name}")
    return sata_devices

def check_device_exists(dev):
    if os.path.exists(dev):
        print(f"Device {dev} found.")
        return True
    print(f"Device {dev} does not exist.")
    return False

def check_secure_erase_supported(dev):
    output = run_cmd(f"hdparm -I {dev}")
    if output and "supported" in output.lower() and ("security erase" in output.lower() or "security erase unit" in output.lower()):
        print("Secure Erase is supported by this device.")
        return True
    print("Secure Erase NOT supported.")
    return False

def is_device_locked(dev):
    output = run_cmd(f"hdparm -I {dev}")
    if output:
        for line in output.splitlines():
            if "not frozen" in line.lower():
                print("Device is NOT frozen.")
                return False
            if "frozen" in line.lower():
                print("Device is frozen. You must power cycle or unplug device before Secure Erase.")
                return True
    return False

def set_password(dev, password="password"):
    print("Setting temporary security password for Secure Erase...")
    out = run_cmd(f"hdparm --user-master u --security-set-pass {password} {dev}")
    if out is None:
        print("Failed to set security password.")
        return False
    print("Password set.")
    return True

def secure_erase(dev, password="password"):
    print("Starting Secure Erase...")
    out = run_cmd(f"hdparm --user-master u --security-erase {password} {dev}")
    if out is None:
        print("Secure Erase failed.")
        return False
    print("Secure Erase command issued successfully.")
    return True

def random_overwrite(dev, passes=3, block_size=1024*1024):
    print(f"[*] Performing {passes}-pass random overwrite (slow)...")
    size_output = run_cmd(f"blockdev --getsize64 {dev}")
    if size_output is None:
        print("Could not get device size.")
        return False
    size = int(size_output)
    print(f"Device size: {size} bytes")
    try:
        with open(dev, "wb") as f:
            for pass_num in range(passes):
                print(f"    Pass {pass_num+1} of {passes}...")
                bytes_written = 0
                while bytes_written < size:
                    data = os.urandom(min(block_size, size - bytes_written))
                    f.write(data)
                    bytes_written += len(data)
                f.flush()
                os.fsync(f.fileno())
        print("Random overwrite complete.")
        return True
    except PermissionError:
        print("Permission denied. Run script as root!")
        return False
    except Exception as e:
        print(f"Error during overwrite: {e}")
        return False

def main():
    if len(sys.argv) == 1:
        print("No device specified. Auto-detecting SATA devices...")
        devices = list_sata_devices()
        if not devices:
            print("No SATA devices found.")
            sys.exit(1)
        print(f"Detected SATA devices: {', '.join(devices)}")
        device = devices[0]
        print(f"Using first SATA device: {device}")
    elif len(sys.argv) == 2:
        device = sys.argv[1]
    else:
        print(f"Usage: sudo python3 {sys.argv[0]} [optional: /dev/sdX]")
        sys.exit(1)

    if not check_device_exists(device):
        sys.exit(1)

    if is_device_locked(device):
        print("Device is frozen. Please power cycle the drive and rerun script.")
        sys.exit(1)

    if check_secure_erase_supported(device):
        if not set_password(device):
            print("Could not set security password; aborting.")
            sys.exit(1)

        if not secure_erase(device):
            print("Secure Erase command failed; aborting.")
            sys.exit(1)

        print("Secure Erase initiated successfully! Wait for completion before unplugging device.")
    else:
        print("Secure Erase not supported. Falling back to multi-pass random overwrite.")
        if not random_overwrite(device):
            print("Failed to overwrite device.")
            sys.exit(1)

    print("Wipe completed.")

if __name__ == "__main__":
    main()
