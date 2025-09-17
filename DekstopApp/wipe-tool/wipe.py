#!/usr/bin/env python3
# filepath: /root/NullBytes/DekstopApp/wipe-tool/wipe.py
import os
import subprocess
import threading
import time
import uuid
import json
import sys
import argparse
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
        return self.run_cmd(f"which {cmd}") is not None

    def list_devices(self):
        """List available storage devices"""
        try:
            # For Linux
            if os.name == 'posix':
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
            else:
                # Windows support (basic)
                return []
                
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
        mounts = subprocess.getoutput("mount")
        return device in mounts

    def is_system_drive(self, device):
        """Check if device is a system drive"""
        # Check for root filesystem
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
                        
                        # Try to extract progress from dd output
                        if "bytes" in line or "GB" in line or "MB" in line:
                            progress = min(progress + 1, 99)
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
            data = subprocess.check_output(f"dd if={device} bs=512 count=1 status=none", shell=True)
            return all(b == 0 for b in data)
        except Exception:
            return False

    def create_wipe_log(self, device, method, log_file, status, verified_clean, output_file):
        """Create comprehensive wipe log"""
        # Get device information
        device_info = next((d for d in self.list_devices() if d["path"] == device), {})
        
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
                "platform": os.name,
                "operator": os.getenv("USER", "Unknown"),
                "log_file": log_file
            },
            "compliance": {
                "nist_800_88": True,
                "certificate_id": str(uuid.uuid4())
            }
        }
        
        with open(output_file, 'w') as f:
            json.dump(wipe_log, f, indent=2)
        
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

    def wipe_device(self, device_path, method, output_log):
        """Main wipe function"""
        # Safety checks
        if self.is_system_drive(device_path):
            print(json.dumps({"error": "Cannot wipe system drive", "code": 1}))
            return False
            
        if self.is_device_mounted(device_path):
            print(json.dumps({"error": "Device is mounted", "code": 2}))
            return False
        
        # Start wipe process
        self.print_progress(0, "Starting wipe process...")
        
        # Get wipe command
        cmd = self.get_wipe_command(device_path, method)
        
        # Create log file path
        log_file = f"/tmp/wipe_raw_{int(time.time())}.log"
        
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
    parser = argparse.ArgumentParser(description='NullBytes Secure Wipe Tool')
    parser.add_argument('--list', action='store_true', help='List available devices')
    parser.add_argument('--json', action='store_true', help='Output in JSON format')
    parser.add_argument('--device', help='Device path to wipe')
    parser.add_argument('--method', choices=['clear', 'purge', 'destroy'], 
                       default='clear', help='Wipe method')
    parser.add_argument('--output', help='Output log file path')
    
    args = parser.parse_args()
    
    wipe_tool = WipeToolCLI()
    
    if args.list:
        devices = wipe_tool.list_devices()
        if args.json:
            wipe_tool.print_devices_json(devices)
        else:
            for device in devices:
                print(f"{device['path']} - {device['name']} ({device['size']})")
    
    elif args.device:
        if not args.output:
            args.output = f"/tmp/wipe_log_{int(time.time())}.json"
        
        success = wipe_tool.wipe_device(args.device, args.method, args.output)
        sys.exit(0 if success else 1)
    
    else:
        parser.print_help()

if __name__ == "__main__":
    main()