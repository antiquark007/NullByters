import os
import shutil
import psutil
import time
from pathlib import Path

def get_usb_drives():
    """
    Detect all USB drives connected to the system
    Returns a list of USB drive letters
    """
    usb_drives = []
    
    # Get all disk partitions
    partitions = psutil.disk_partitions()
    
    for partition in partitions:
        # Check if it's a removable drive (USB)
        if 'removable' in partition.opts:
            usb_drives.append(partition.mountpoint)
    
    return usb_drives

def delete_usb_contents(drive_path):
    """
    Delete all contents from the specified USB drive
    """
    try:
        if not os.path.exists(drive_path):
            print(f"Drive {drive_path} not found!")
            return False
        
        print(f"Deleting contents from {drive_path}...")
        
        # Get all items in the USB drive
        for item in os.listdir(drive_path):
            item_path = os.path.join(drive_path, item)
            
            try:
                if os.path.isfile(item_path) or os.path.islink(item_path):
                    # Delete file or symbolic link
                    os.unlink(item_path)
                    print(f"Deleted file: {item}")
                elif os.path.isdir(item_path):
                    # Delete directory and all its contents
                    shutil.rmtree(item_path)
                    print(f"Deleted directory: {item}")
            except Exception as e:
                print(f"Failed to delete {item}: {str(e)}")
        
        print(f"Successfully cleaned USB drive: {drive_path}")
        return True
        
    except PermissionError:
        print(f"Permission denied! Run as administrator to delete from {drive_path}")
        return False
    except Exception as e:
        print(f"Error deleting contents from {drive_path}: {str(e)}")
        return False

def monitor_and_clean_usb():
    """
    Continuously monitor for USB drives and clean them when detected
    """
    print("USB Monitor started. Waiting for USB drives...")
    print("Press Ctrl+C to stop monitoring")
    
    known_drives = set()
    
    try:
        while True:
            current_drives = set(get_usb_drives())
            
            # Check for newly connected USB drives
            new_drives = current_drives - known_drives
            
            for drive in new_drives:
                print(f"\nNew USB drive detected: {drive}")
                
                # Ask for confirmation before deleting
                response = input(f"Delete all contents from {drive}? (y/N): ").strip().lower()
                
                if response == 'y' or response == 'yes':
                    delete_usb_contents(drive)
                else:
                    print(f"Skipping deletion for {drive}")
            
            # Update known drives
            known_drives = current_drives
            
            # Wait before checking again
            time.sleep(2)
            
    except KeyboardInterrupt:
        print("\nUSB monitoring stopped.")

def list_usb_drives():
    """
    List all currently connected USB drives
    """
    drives = get_usb_drives()
    
    if drives:
        print("Connected USB drives:")
        for i, drive in enumerate(drives, 1):
            try:
                usage = shutil.disk_usage(drive)
                total_gb = usage.total / (1024**3)
                used_gb = usage.used / (1024**3)
                free_gb = usage.free / (1024**3)
                
                print(f"{i}. {drive}")
                print(f"   Total: {total_gb:.2f} GB")
                print(f"   Used: {used_gb:.2f} GB")
                print(f"   Free: {free_gb:.2f} GB")
            except Exception as e:
                print(f"{i}. {drive} (Error reading disk info: {e})")
    else:
        print("No USB drives detected.")
    
    return drives

def main():
    """
    Main function with menu options
    """
    print("=== USB Drive Manager ===")
    print("1. List USB drives")
    print("2. Delete contents from specific USB drive")
    print("3. Monitor and auto-clean USB drives")
    print("4. Exit")
    
    while True:
        try:
            choice = input("\nEnter your choice (1-4): ").strip()
            
            if choice == '1':
                list_usb_drives()
                
            elif choice == '2':
                drives = list_usb_drives()
                if drives:
                    try:
                        drive_num = int(input(f"Select drive (1-{len(drives)}): ")) - 1
                        if 0 <= drive_num < len(drives):
                            selected_drive = drives[drive_num]
                            
                            # Double confirmation for safety
                            confirm = input(f"Are you sure you want to delete ALL contents from {selected_drive}? Type 'DELETE' to confirm: ")
                            if confirm == 'DELETE':
                                delete_usb_contents(selected_drive)
                            else:
                                print("Operation cancelled.")
                        else:
                            print("Invalid selection.")
                    except ValueError:
                        print("Please enter a valid number.")
                        
            elif choice == '3':
                monitor_and_clean_usb()
                
            elif choice == '4':
                print("Goodbye!")
                break
                
            else:
                print("Invalid choice. Please enter 1-4.")
                
        except KeyboardInterrupt:
            print("\nGoodbye!")
            break

if __name__ == "__main__":
    main()