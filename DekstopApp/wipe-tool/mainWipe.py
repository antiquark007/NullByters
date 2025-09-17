#!/usr/bin/env python3
# filepath: /root/NullBytes/DekstopApp/wipe-tool/mainWipe.py
"""
NullBytes Secure Wipe Tool - Cross Platform Entry Point
Automatically detects the operating system and uses the appropriate wipe implementation.
"""
import os
import sys
import json
import argparse
from datetime import datetime

def get_platform():
    """Detect the operating system"""
    if os.name == 'nt':
        return 'windows'
    elif os.name == 'posix':
        return 'linux'
    else:
        return 'unknown'

def import_platform_module():
    """Import the appropriate platform-specific module"""
    platform = get_platform()
    
    try:
        if platform == 'windows':
            from windowsWipe import WindowsWipeToolCLI
            return WindowsWipeToolCLI(), 'Windows'
        elif platform == 'linux':
            from linuxWipe import LinuxWipeToolCLI
            return LinuxWipeToolCLI(), 'Linux'
        else:
            raise ImportError(f"Unsupported platform: {platform}")
    except ImportError as e:
        print(json.dumps({
            "error": f"Failed to import platform module: {str(e)}",
            "platform": platform,
            "code": 4
        }))
        sys.exit(1)

def print_platform_info():
    """Print platform information"""
    platform = get_platform()
    platform_info = {
        "platform": platform,
        "os_name": os.name,
        "timestamp": datetime.now().isoformat(),
        "python_version": f"{sys.version_info.major}.{sys.version_info.minor}.{sys.version_info.micro}"
    }
    
    if platform == 'windows':
        try:
            import platform as plt
            platform_info["windows_version"] = f"{plt.system()} {plt.release()}"
        except:
            platform_info["windows_version"] = "Unknown"
    elif platform == 'linux':
        try:
            import platform as plt
            platform_info["linux_distribution"] = plt.platform()
            with open('/etc/os-release', 'r') as f:
                for line in f:
                    if line.startswith('PRETTY_NAME='):
                        platform_info["distribution"] = line.split('=')[1].strip().strip('"')
                        break
        except:
            platform_info["distribution"] = "Unknown Linux"
    
    print(json.dumps(platform_info, indent=2))

def main():
    parser = argparse.ArgumentParser(description='NullBytes Secure Wipe Tool - Cross Platform')
    parser.add_argument('--list', action='store_true', help='List available devices')
    parser.add_argument('--json', action='store_true', help='Output in JSON format')
    parser.add_argument('--device', help='Device path to wipe')
    parser.add_argument('--method', choices=['clear', 'purge', 'destroy'], 
                       default='clear', help='Wipe method (clear=1 pass, purge=3 pass, destroy=7 pass)')
    parser.add_argument('--output', help='Output log file path')
    parser.add_argument('--check-admin', action='store_true', help='Check if running with admin/root privileges')
    parser.add_argument('--platform-info', action='store_true', help='Show platform information')
    parser.add_argument('--version', action='version', version='NullBytes Wipe Tool v1.0.0')
    
    args = parser.parse_args()
    
    # Handle platform info request
    if args.platform_info:
        print_platform_info()
        return
    
    # Import platform-specific module
    wipe_tool, platform_name = import_platform_module()
    
    # Handle admin check
    if args.check_admin:
        if platform_name == 'Windows':
            is_admin = wipe_tool.check_admin_privileges()
        else:  # Linux
            is_admin = wipe_tool.check_root_privileges()
        
        print(json.dumps({
            "admin": is_admin, 
            "platform": platform_name,
            "required_privilege": "Administrator" if platform_name == 'Windows' else "Root"
        }))
        return
    
    # Handle device listing
    if args.list:
        devices = wipe_tool.list_devices()
        if args.json:
            wipe_tool.print_devices_json(devices)
        else:
            print(f"\n=== Available Storage Devices on {platform_name} ===")
            if not devices:
                print("No storage devices found.")
            else:
                for device in devices:
                    removable_str = " (Removable)" if device.get('removable', False) else ""
                    device_type = device.get('device_type', 'Unknown')
                    print(f"{device['path']} - {device['name']} ({device['size']}) - {device_type}{removable_str}")
            print(f"\nTotal devices found: {len(devices)}")
    
    # Handle device wiping
    elif args.device:
        if not args.output:
            # Set platform-appropriate default output path
            if platform_name == 'Windows':
                temp_dir = os.environ.get('TEMP', 'C:\\temp')
                args.output = os.path.join(temp_dir, f"wipe_log_{int(time.time())}.json")
            else:  # Linux
                args.output = f"/tmp/wipe_log_{int(time.time())}.json"
        
        print(json.dumps({
            "message": f"Starting {platform_name} wipe operation",
            "device": args.device,
            "method": args.method,
            "output": args.output,
            "timestamp": datetime.now().isoformat()
        }))
        
        success = wipe_tool.wipe_device(args.device, args.method, args.output)
        
        if success:
            print(json.dumps({
                "message": f"{platform_name} wipe completed successfully",
                "log_file": args.output,
                "timestamp": datetime.now().isoformat()
            }))
        
        sys.exit(0 if success else 1)
    
    else:
        print(f"NullBytes Secure Wipe Tool v1.0.0 - {platform_name}")
        print("=" * 50)
        parser.print_help()
        print(f"\nDetected Platform: {platform_name}")
        print(f"Operating System: {os.name}")
        print("\nSafety Features:")
        print("- System drive protection")
        print("- Mount point checking")
        print("- Administrative privilege verification")
        print("- NIST 800-88 compliance")

if __name__ == "__main__":
    main()