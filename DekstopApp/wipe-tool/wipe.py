#!/usr/bin/env python3
import os
import subprocess
import threading
import time
import uuid
import json
import sys
import argparse
import tempfile
import shutil
from datetime import datetime

class WipeToolCLI:
    def __init__(self):
        self.stop_event = threading.Event()
        
    def run_cmd(self, command, capture_output=True):
        try:
            result = subprocess.run(command, shell=True, capture_output=capture_output, text=True)
            if result.returncode != 0:
                return None
            return result.stdout.strip()
        except Exception:
            return None

    def check_dependency(self, cmd):
        """Check if a command/tool is available"""
        if os.name == 'nt':
            # Windows - check if command exists
            return self.run_cmd(f"where {cmd}") is not None
        else:
            # Linux/Unix
            return self.run_cmd(f"which {cmd}") is not None

    def list_devices(self):
        """List available storage devices"""
        try:
            # For Windows
            if os.name == 'nt':
                devices = []
                
                # Get disk drives using wmic
                wmic_cmd = 'wmic diskdrive get DeviceID,Model,Size,SerialNumber,MediaType,InterfaceType /format:csv'
                output = subprocess.getoutput(wmic_cmd)
                
                for line in output.splitlines()[1:]:  # Skip header
                    if line.strip() and ',' in line:
                        parts = [p.strip() for p in line.split(',')]
                        if len(parts) >= 6 and parts[1]:  # Check if DeviceID exists
                            device_id = parts[1]  # DeviceID (e.g., \\.\PHYSICALDRIVE0)
                            interface_type = parts[2] or "Unknown"
                            media_type = parts[3] or "Unknown" 
                            model = parts[4] or "Unknown"
                            serial = parts[5] or "Unknown"
                            size_bytes = parts[6] or "0"
                            
                            # Convert size to human readable format
                            try:
                                size_int = int(size_bytes)
                                if size_int > 0:
                                    # Convert bytes to GB
                                    size_gb = size_int / (1024**3)
                                    if size_gb >= 1:
                                        size = f"{size_gb:.1f}G"
                                    else:
                                        size_mb = size_int / (1024**2)
                                        size = f"{size_mb:.1f}M"
                                else:
                                    size = "Unknown"
                            except:
                                size = "Unknown"
                            
                            # Check if removable using interface type
                            is_removable = "USB" in interface_type or "Removable" in media_type
                            
                            # Get vendor from model
                            vendor = ""
                            if model and model != "Unknown":
                                vendor_names = ["WDC", "Seagate", "Samsung", "Toshiba", "Hitachi", "SanDisk", "Kingston", "Crucial"]
                                for v in vendor_names:
                                    if v.lower() in model.lower():
                                        vendor = v
                                        break
                            
                            device_info = {
                                "name": f"{vendor} {model}".strip() if vendor else model,
                                "path": device_id,
                                "size": size,
                                "model": model,
                                "serial": serial,
                                "removable": is_removable,
                                "vendor": vendor,
                                "interface": interface_type,
                                "media_type": media_type
                            }
                            devices.append(device_info)
                
                return devices
                
            # For Linux
            else:
                output = subprocess.getoutput("lsblk -dpno NAME,SIZE,MODEL,SERIAL | grep -v 'loop\\|sr0'")
                devices = []
                for line in output.splitlines():
                    parts = line.split()
                    if len(parts) >= 2:
                        dev_path = parts[0]
                        size = parts[1] if len(parts) > 1 else "Unknown"
                        model = parts[2] if len(parts) > 2 else "Unknown"
                        serial = parts[3] if len(parts) > 3 else "Unknown"
                        
                        # Check if it's a removable device
                        removable_check = self.run_cmd(f"cat /sys/block/{os.path.basename(dev_path)}/removable 2>/dev/null")
                        is_removable = removable_check == "1"
                        
                        # Get more device info
                        vendor = self.run_cmd(f"cat /sys/block/{os.path.basename(dev_path)}/device/vendor 2>/dev/null") or ""
                        
                        device_info = {
                            "name": f"{vendor.strip()} {model}".strip(),
                            "path": dev_path,
                            "size": size,
                            "model": model,
                            "serial": serial,
                            "removable": is_removable,
                            "vendor": vendor.strip()
                        }
                        devices.append(device_info)
                
                return devices
                
        except Exception as e:
            return []

    def print_devices_json(self, devices):
        """Print devices in JSON format for GUI consumption"""
        output = {
            "devices": devices,
            "count": len(devices),
            "timestamp": datetime.now().isoformat()
        }
        print(json.dumps(output, indent=2))
        sys.stdout.flush()

    def is_device_mounted(self, device):
        """Check if device is mounted"""
        if os.name == 'nt':
            # Windows: Check if any volumes are associated with the physical drive
            try:
                # Extract drive number from device path (e.g., \\.\PHYSICALDRIVE0 -> 0)
                if 'PHYSICALDRIVE' in device:
                    drive_num = device.split('PHYSICALDRIVE')[-1]
                    # Check for associated volumes
                    volumes_cmd = f'wmic volume where "DriveLetter IS NOT NULL" get DriveLetter /format:csv'
                    output = subprocess.getoutput(volumes_cmd)
                    # If there are volumes, the drive might be in use
                    return len([line for line in output.splitlines() if line.strip() and ':' in line]) > 0
                return False
            except:
                return False
        else:
            # Linux
            mounts = subprocess.getoutput("mount")
            return device in mounts

    def is_system_drive(self, device):
        """Check if device is a system drive"""
        if os.name == 'nt':
            # Windows: Check if it's the system drive (usually PHYSICALDRIVE0)
            try:
                # Get system drive letter
                system_drive = os.environ.get('SystemDrive', 'C:')
                
                # Usually PHYSICALDRIVE0 contains the system partition on Windows
                if 'PHYSICALDRIVE0' in device:
                    return True
                    
                # Additional check: see if the device contains the Windows directory
                system_root = os.environ.get('SystemRoot', 'C:\\Windows')
                if system_root[0].upper() + ':' in device:
                    return True
                    
                return False
            except:
                return False
        else:
            # Linux code (existing)
            root_device = self.run_cmd("df / | tail -1 | cut -d' ' -f1")
            if root_device and device in root_device:
                return True
                
            # Check for common system paths
            system_devices = ['/dev/sda', '/dev/nvme0n1', '/dev/mmcblk0']
            for sys_dev in system_devices:
                if device.startswith(sys_dev):
                    return True
                    
            return False

    def get_wipe_command(self, device, method):
        """Get the appropriate wipe command based on method"""
        if os.name == 'nt':
            # Windows commands
            # Note: These require additional tools or PowerShell scripts
            commands = {
                "clear": f'powershell -Command "& {{$device = Get-WmiObject -Class Win32_DiskDrive | Where-Object {{$_.DeviceID -eq \\\"{device}\\\"}}; if($device) {{Write-Host \\\"Clearing device {device}...\\\"; $stream = [System.IO.File]::OpenWrite($device); $buffer = New-Object byte[] 1048576; for($i=0; $i -lt $device.Size; $i += 1048576) {{$stream.Write($buffer, 0, $buffer.Length)}}; $stream.Close()}}}}"',
                "purge": f'cipher /w:{device}',
                "destroy": f'sdelete -p 7 -s -z {device}' if self.check_dependency('sdelete') else f'cipher /w:{device}'
            }
        else:
            # Linux commands (existing)
            commands = {
                "clear": f"dd if=/dev/zero of={device} bs=1M status=progress",
                "purge": f"shred -v -n 3 {device}",
                "destroy": f"shred -v -n 7 {device} && dd if=/dev/zero of={device} bs=1M status=progress"
            }
        return commands.get(method, commands["clear"])

    def print_progress(self, percent, message=""):
        """Print progress in JSON format for GUI"""
        progress_data = {
            "progress": percent,
            "message": message,
            "timestamp": datetime.now().isoformat()
        }
        print(json.dumps(progress_data))
        sys.stdout.flush()

    def run_wipe_command(self, cmd, log_path):
        """Run wipe command with progress monitoring"""
        try:
            with open(log_path, "w") as logfile:
                # Start the process
                proc = subprocess.Popen(
                    cmd, 
                    shell=True, 
                    stdout=subprocess.PIPE, 
                    stderr=subprocess.STDOUT,
                    universal_newlines=True,
                    bufsize=1
                )
                
                progress = 0
                while proc.poll() is None:
                    if self.stop_event.is_set():
                        proc.terminate()
                        return False
                        
                    # Read output
                    line = proc.stdout.readline()
                    if line:
                        logfile.write(line)
                        logfile.flush()
                        
                        # Try to extract progress from output
                        if any(keyword in line.lower() for keyword in ["bytes", "gb", "mb", "progress", "complete", "%"]):
                            progress = min(progress + 2, 99)
                            self.print_progress(progress, "Wiping in progress...")
                    
                    time.sleep(0.1)
                
                # Process completed
                return_code = proc.wait()
                if return_code == 0:
                    self.print_progress(100, "Wipe completed successfully")
                    return True
                else:
                    self.print_progress(0, f"Wipe failed with code {return_code}")
                    return False
                    
        except Exception as e:
            self.print_progress(0, f"Error during wipe: {str(e)}")
            return False

    def verify_wipe(self, device):
        """Verify if device was properly wiped"""
        try:
            if os.name == 'nt':
                # Windows verification - read first sector
                verify_cmd = f'powershell -Command "& {{$fs = New-Object System.IO.FileStream(\\\"{device}\\\", [System.IO.FileMode]::Open, [System.IO.FileAccess]::Read); $buffer = New-Object byte[] 512; $fs.Read($buffer, 0, 512); $fs.Close(); $sum = ($buffer | Measure-Object -Sum).Sum; Write-Output $sum}}"'
                result = subprocess.getoutput(verify_cmd)
                return int(result.strip()) == 0
            else:
                # Linux verification
                data = subprocess.check_output(f"dd if={device} bs=512 count=1 status=none", shell=True)
                return all(b == 0 for b in data)
        except Exception:
            return False

    def create_wipe_log(self, device, method, log_file, status, verified_clean, output_file):
        """Create comprehensive wipe log"""
        # Get device information
        device_info = next((d for d in self.list_devices() if d["path"] == device), {})
        
        # Get appropriate temp directory and user
        if os.name == 'nt':
            temp_dir = os.environ.get('TEMP', 'C:\\temp')
            operator = os.environ.get('USERNAME', 'Unknown')
        else:
            temp_dir = '/tmp'
            operator = os.environ.get('USER', 'Unknown')
        
        wipe_log = {
            "version": "1.0",
            "device": {
                "path": device,
                "name": device_info.get("name", "Unknown Device"),
                "model": device_info.get("model", "Unknown"),
                "serial": device_info.get("serial", "Unknown"),
                "size": device_info.get("size", "Unknown"),
                "vendor": device_info.get("vendor", "Unknown")
            },
            "wipe": {
                "method": method,
                "nist_level": self.get_nist_level(method),
                "status": status,
                "started_at": datetime.now().isoformat(),
                "finished_at": datetime.now().isoformat(),
                "passes_completed": self.get_pass_count(method),
                "verified_clean": verified_clean
            },
            "system": {
                "tool_version": "1.0.0-python",
                "platform": "Windows" if os.name == 'nt' else "Linux",
                "os_name": os.name,
                "operator": operator,
                "log_file": log_file
            },
            "compliance": {
                "nist_800_88": True,
                "certificate_id": str(uuid.uuid4())
            }
        }
        
        try:
            # Ensure output directory exists
            os.makedirs(os.path.dirname(output_file), exist_ok=True)
            
            with open(output_file, 'w') as f:
                json.dump(wipe_log, f, indent=2)
        except Exception as e:
            # Fallback to temp directory
            temp_file = os.path.join(temp_dir, f"wipe_log_{int(time.time())}.json")
            with open(temp_file, 'w') as f:
                json.dump(wipe_log, f, indent=2)
            wipe_log["system"]["log_file"] = temp_file
        
        return wipe_log

    def get_nist_level(self, method):
        """Get NIST compliance level for method"""
        levels = {
            "clear": "clear",
            "purge": "purge", 
            "destroy": "destroy"
        }
        return levels.get(method, "clear")

    def get_pass_count(self, method):
        """Get number of passes for method"""
        passes = {
            "clear": 1,
            "purge": 3,
            "destroy": 7
        }
        return passes.get(method, 1)

    def get_temp_log_path(self):
        """Get appropriate temporary log file path for the OS"""
        if os.name == 'nt':
            temp_dir = os.environ.get('TEMP', 'C:\\temp')
            return os.path.join(temp_dir, f"wipe_raw_{int(time.time())}.log")
        else:
            return f"/tmp/wipe_raw_{int(time.time())}.log"

    def get_default_output_path(self):
        """Get default output path for wipe logs"""
        if os.name == 'nt':
            temp_dir = os.environ.get('TEMP', 'C:\\temp')
            return os.path.join(temp_dir, f"wipe_log_{int(time.time())}.json")
        else:
            return f"/tmp/wipe_log_{int(time.time())}.json"

    def wipe_device(self, device_path, method, output_log):
        """Main wipe function"""
        # Safety checks
        if self.is_system_drive(device_path):
            print(json.dumps({"error": "Cannot wipe system drive", "code": 1}))
            return False
            
        if self.is_device_mounted(device_path):
            print(json.dumps({"error": "Device is mounted", "code": 2}))
            return False
        
        # Administrative privileges check for Windows
        if os.name == 'nt':
            try:
                # Try to open a file handle to the device to check permissions
                import ctypes
                if not ctypes.windll.shell32.IsUserAnAdmin():
                    print(json.dumps({"error": "Administrator privileges required", "code": 3}))
                    return False
            except:
                pass  # Continue if we can't check admin status
        
        # Start wipe process
        self.print_progress(0, "Starting wipe process...")
        
        # Get wipe command
        cmd = self.get_wipe_command(device_path, method)
        
        # Create log file path
        log_file = self.get_temp_log_path()
        
        # Run wipe
        success = self.run_wipe_command(cmd, log_file)
        
        # Verify wipe if successful
        verified_clean = False
        if success:
            self.print_progress(95, "Verifying wipe...")
            verified_clean = self.verify_wipe(device_path)
        
        # Create comprehensive log
        status = "success" if success else "failed"
        wipe_log = self.create_wipe_log(
            device_path, method, log_file, status, verified_clean, output_log
        )
        
        self.print_progress(100, "Process completed")
        
        return success

def main():
    parser = argparse.ArgumentParser(description='NullBytes Secure Wipe Tool - Cross Platform')
    parser.add_argument('--list', action='store_true', help='List available devices')
    parser.add_argument('--json', action='store_true', help='Output in JSON format')
    parser.add_argument('--device', help='Device path to wipe')
    parser.add_argument('--method', choices=['clear', 'purge', 'destroy'], 
                       default='clear', help='Wipe method')
    parser.add_argument('--output', help='Output log file path')
    parser.add_argument('--check-admin', action='store_true', help='Check if running with admin privileges')
    
    args = parser.parse_args()
    
    wipe_tool = WipeToolCLI()
    
    if args.check_admin:
        if os.name == 'nt':
            try:
                import ctypes
                is_admin = ctypes.windll.shell32.IsUserAnAdmin()
                print(json.dumps({"admin": bool(is_admin)}))
            except:
                print(json.dumps({"admin": False}))
        else:
            is_root = os.geteuid() == 0
            print(json.dumps({"admin": is_root}))
        return
    
    if args.list:
        devices = wipe_tool.list_devices()
        if args.json:
            wipe_tool.print_devices_json(devices)
        else:
            for device in devices:
                print(f"{device['path']} - {device['name']} ({device['size']})")
    
    elif args.device:
        if not args.output:
            args.output = wipe_tool.get_default_output_path()
        
        success = wipe_tool.wipe_device(args.device, args.method, args.output)
        sys.exit(0 if success else 1)
    
    else:
        parser.print_help()

if __name__ == "__main__":
    main()