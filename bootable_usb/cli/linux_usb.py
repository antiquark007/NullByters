#!/usr/bin/env python3
import os
import subprocess
import time

def get_usb_devices():
    """Find USB devices using multiple methods"""
    usb_devices = []
    
    try:
        # Use lsblk to find removable devices
        result = subprocess.run(['lsblk', '-d', '-o', 'NAME,RO,RM,SIZE,TYPE'], 
                              capture_output=True, text=True, check=True)
        
        for line in result.stdout.splitlines()[1:]:  # Skip header
            parts = line.split()
            if len(parts) >= 5 and parts[4] == 'disk' and parts[2] == '1':  # RM=1 means removable
                device_name = '/dev/' + parts[0]
                # Only add the device, not partitions
                if not any(char.isdigit() for char in parts[0]):
                    usb_devices.append(device_name)
                        
    except Exception as e:
        print(f"Error finding USB devices: {e}")
        
    return usb_devices

def unmount_device(device):
    """Unmount all partitions of a device"""
    try:
        # Find all partitions of this device
        partitions = []
        for part in os.listdir('/dev'):
            if part.startswith(device.split('/')[-1]) and part != device.split('/')[-1]:
                partitions.append('/dev/' + part)
        
        # Unmount all partitions
        for partition in partitions:
            try:
                print(f"Unmounting {partition}...")
                subprocess.run(['umount', '-f', partition], check=True, timeout=10)
            except subprocess.TimeoutExpired:
                print(f"Force unmounting {partition}...")
                subprocess.run(['umount', '-l', partition], check=True)
            except Exception as e:
                print(f"Could not unmount {partition}: {e}")
        
        # Wait a moment for unmount to complete
        time.sleep(2)
        return True
    except Exception as e:
        print(f"Error unmounting device {device}: {e}")
        return False

def quick_wipe_device(device):
    """Quickly wipe a USB device by destroying filesystem structures"""
    try:
        print(f"Quickly wiping {device}...")
        
        # Unmount all partitions first
        if not unmount_device(device):
            print(f"Failed to unmount {device}, trying to continue...")
        
        # Method 1: Use wipefs to remove all filesystem signatures
        print("Removing filesystem signatures...")
        subprocess.run(['wipefs', '-a', device], check=True, timeout=30)
        
        # Method 2: Zero out the first and last 1MB of the device
        # This destroys partition tables and filesystem metadata
        print("Destroying partition tables...")
        
        # Get device size
        result = subprocess.run(['blockdev', '--getsize64', device], 
                              capture_output=True, text=True, check=True)
        device_size = int(result.stdout.strip())
        
        # Zero out the beginning (first 10MB)
        subprocess.run(['dd', 'if=/dev/zero', f'of={device}', 'bs=1M', 'count=10'], 
                      check=True, timeout=30)
        
        # Zero out the end (last 1MB)
        if device_size > 1048576:  # Only if device is larger than 1MB
            seek_position = (device_size - 1048576) // 1048576
            subprocess.run(['dd', 'if=/dev/zero', f'of={device}', 'bs=1M', 
                          f'seek={seek_position}', 'count=1'], check=True, timeout=30)
        
        # Create a new partition table (MBR)
        print("Creating new partition table...")
        subprocess.run(['parted', '-s', device, 'mklabel', 'msdos'], check=True, timeout=10)
        
        # Create a new primary partition using the whole disk
        print("Creating new partition...")
        subprocess.run(['parted', '-s', device, 'mkpart', 'primary', 'fat32', '0%', '100%'], 
                      check=True, timeout=10)
        
        # Wait for partition to be recognized
        time.sleep(2)
        
        # Find the partition name
        partition = None
        for part in os.listdir('/dev'):
            if part.startswith(device.split('/')[-1]) and part != device.split('/')[-1]:
                partition = '/dev/' + part
                break
        
        if not partition:
            print(f"Could not find partition for {device}")
            return False
        
        # Format the partition
        print(f"Formatting {partition}...")
        subprocess.run(['mkfs.vfat', '-F', '32', '-n', 'USBDRIVE', partition], 
                      check=True, timeout=30)
        
        print(f"Successfully wiped and formatted {device}")
        return True
        
    except subprocess.TimeoutExpired:
        print(f"Timeout while processing {device}")
        return False
    except Exception as e:
        print(f"Error wiping {device}: {e}")
        return False

def main():
    print("Finding USB devices...")
    devices = get_usb_devices()
    
    if not devices:
        print("No USB devices found")
        return
        
    print(f"Found USB devices: {devices}")
    
    for device in devices:
        # Double confirmation for each device
        confirm = input(f"Type 'YES' to wipe {device} (THIS WILL DESTROY ALL DATA): ")
        if confirm == 'YES':
            success = quick_wipe_device(device)
            if success:
                print(f"Successfully processed {device}")
            else:
                print(f"Failed to process {device}")
        else:
            print(f"Skipping {device}")

if __name__ == '__main__':
    main()
