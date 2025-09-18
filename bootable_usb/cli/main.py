import os
import subprocess
import sys

def main():
    # Detect the operating system
    if os.name == 'posix':
        print("Linux detected")
        # Ask for confirmation
        confirm = input("Type 123 to confirm USB content deletion: ")
        if confirm == '123':
            # Run Linux USB script with sudo
            script_path = os.path.join(os.path.dirname(__file__), 'linux_usb.py')
            subprocess.run(['sudo', 'python3', script_path])
        else:
            print("Deletion cancelled")
    elif os.name == 'nt':
        print("Windows detected - calling windows_usb.py")
        # Add Windows implementation here (to be done by friend)
    else:
        print("Unsupported operating system")

if __name__ == '__main__':
    main()