import subprocess
import sys
import time

def list_nvme_devices():
    try:
        result = subprocess.run(
            ["lsblk", "-dpno", "NAME,TRAN,SIZE,TYPE,MOUNTPOINT"],
            stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True, check=True
        )

        devices = [line for line in result.stdout.splitlines() if "nvme" in line]

        if not devices:
            print("No NVMe devices found.")
            sys.exit(1)

        print("\nDetected NVMe Devices:")
        for idx, device in enumerate(devices, 1):
            print(f"{idx}. {device}")

        return devices

    except subprocess.CalledProcessError as e:
        print(f"Error occurred while listing devices: {e.stderr}")
        sys.exit(1)

def select_device(devices):
    while True:
        try:
            choice = int(input("\nEnter the number of the NVMe device to delete: "))
            if 1 <= choice <= len(devices):
                return devices[choice - 1].split()[0]  # Extract the device name (e.g., /dev/nvme0n1)
            else:
                print("Invalid choice, please select a valid number.")
        except ValueError:
            print("Invalid input. Please enter a valid number.")

def verify_device(device):
    print(f"\nVerifying {device}...")
    try:
        result = subprocess.run(
            ["lsblk", "-dpno", "NAME,TRAN,MOUNTPOINT", device],
            stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True, check=True
        )

        if result.stdout.strip():
            print(f"Device {device} is accessible and ready for operation.")
        else:
            print(f"Warning: Device {device} might not be correctly recognized.")
            sys.exit(1)
    except subprocess.CalledProcessError as e:
        print(f"Error verifying device: {e.stderr}")
        sys.exit(1)

def sanitize_device(device):
    print(f"\nSanitizing {device}...")
    try:
        subprocess.run(
            ["sudo", "nvme", "sanitize", device, "--sanact=1"],
            check=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE
        )
        print(f"Sanitization of {device} completed.")
    except subprocess.CalledProcessError as e:
        print(f"Error occurred during sanitization: {e.stderr}")
        sys.exit(1)

def format_device(device):
    print(f"\nFormatting {device}...")
    try:
        subprocess.run(
            ["sudo", "nvme", "format", device, "--ses=1"],
            check=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE
        )
        print(f"Formatting of {device} completed.")
    except subprocess.CalledProcessError as e:
        print(f"Error occurred during formatting: {e.stderr}")
        sys.exit(1)

def confirm_operation():
    confirmation = input("\nAre you sure you want to continue (y/n)? ").strip().lower()
    if confirmation != 'y':
        print("Operation aborted.")
        sys.exit(0)

def main():
    print("Welcome to the NVMe device management tool!")

    devices = list_nvme_devices()

    selected_device = select_device(devices)

    verify_device(selected_device)

    confirm_operation()

    sanitize_device(selected_device)

    format_device(selected_device)

    print("\nOperation completed successfully!")

if __name__ == "__main__":
    main()
