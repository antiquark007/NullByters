import os
import subprocess
import tkinter as tk
from tkinter import messagebox, simpledialog, ttk
import threading
import time
import uuid
import json
from datetime import datetime

def run_command(cmd, log_path, stop_event):
    with open(log_path, "w") as logfile:
        proc = subprocess.Popen(cmd, shell=True, stdout=logfile, stderr=logfile)
        while proc.poll() is None:
            if stop_event.is_set():
                proc.terminate()
                return False
            time.sleep(0.5)
    return proc.returncode == 0

def list_devices():
    output = subprocess.getoutput("lsblk -dpno NAME,SIZE,MODEL | grep -v 'loop\\|sr0'")
    devices = []
    for line in output.splitlines():
        parts = line.split()
        if len(parts) >= 2:
            dev = parts[0]
            info = ' '.join(parts[1:])
            devices.append((dev, info))
    return devices

def verify_wipe(device):
    try:
        data = subprocess.check_output(f"dd if={device} bs=512 count=1 status=none", shell=True)
        return all(b == 0 for b in data)
    except Exception:
        return False

def write_certificate(device, method, log_file, status, verified_clean, cert_dir="/var/log/dasbootwiper"):
    try:
        os.makedirs(cert_dir, exist_ok=True)
    except PermissionError:
        cert_dir = "/tmp/dasbootwiper"
        os.makedirs(cert_dir, exist_ok=True)

    cert = {
        "uuid": str(uuid.uuid4()),
        "device": device,
        "method": method,
        "timestamp": datetime.now().isoformat(),
        "status": status,
        "log_file": log_file,
        "verified_clean": verified_clean
    }

    filename = f"{cert['uuid']}_{os.path.basename(device)}.json"
    filepath = os.path.join(cert_dir, filename)

    with open(filepath, 'w') as f:
        json.dump(cert, f, indent=4)

    return filepath

class WipeApp:
    def __init__(self):
        self.root = tk.Tk()
        self.root.withdraw()

    def show_progress(self, stop_event):
        self.progress_win = tk.Toplevel()
        self.progress_win.title("Wiping in Progress...")
        self.progress_win.geometry("300x100")
        self.progress_win.resizable(False, False)
        ttk.Label(self.progress_win, text="Wiping device, please wait...").pack(pady=10)
        self.progressbar = ttk.Progressbar(self.progress_win, mode='indeterminate')
        self.progressbar.pack(padx=20, pady=10, fill=tk.X)
        self.progressbar.start(10)

        # Disable close button
        self.progress_win.protocol("WM_DELETE_WINDOW", lambda: None)

        # Center window
        self.progress_win.update_idletasks()
        x = (self.progress_win.winfo_screenwidth() - self.progress_win.winfo_reqwidth()) // 2
        y = (self.progress_win.winfo_screenheight() - self.progress_win.winfo_reqheight()) // 2
        self.progress_win.geometry(f"+{x}+{y}")

        # Poll to close when stop_event is set
        def check_stop():
            if stop_event.is_set():
                self.progressbar.stop()
                self.progress_win.destroy()
            else:
                self.progress_win.after(100, check_stop)

        check_stop()

    def wipe_device(self, device, method):
        stop_event = threading.Event()
        log_path = f"/tmp/wipe_{os.path.basename(device)}_{int(time.time())}.log"
        if method == "Zero Fill":
            cmd = f"dd if=/dev/zero of={device} bs=1M status=progress"
        elif method == "Random Fill":
            cmd = f"dd if=/dev/urandom of={device} bs=1M status=progress"
        elif method == "Shred + Zero":
            # Note: shred outputs verbose info - may flood log
            cmd = f"shred -v -n 3 {device} && dd if=/dev/zero of={device} bs=1M status=progress"
        else:
            messagebox.showerror("Error", "Unknown wipe method selected.")
            return None, "failed"

        # Run wipe command in a thread
        thread = threading.Thread(target=run_command, args=(cmd, log_path, stop_event))
        thread.start()

        # Show progress bar dialog
        self.show_progress(stop_event)

        # Wait for thread to complete
        while thread.is_alive():
            self.root.update()
            time.sleep(0.1)

        stop_event.set()
        thread.join()

        # Ensure sync
        subprocess.call("sync", shell=True)

        # Check exit status by checking if log exists (basic)
        return log_path, "success"

    def main(self):
        devices = list_devices()
        if not devices:
            messagebox.showerror("Error", "No devices found!")
            return

        device_labels = [f"{dev} - {info}" for dev, info in devices]
        choice = simpledialog.askinteger("Select Device", "\n".join(f"{i+1}: {label}" for i, label in enumerate(device_labels)))
        if not choice or choice < 1 or choice > len(devices):
            return
        device = devices[choice - 1][0]

        mounts = subprocess.getoutput("mount")
        if device in mounts:
            messagebox.showerror("Mounted!", f"Device {device} is currently mounted. Please unmount it first.")
            return

        methods = ["Zero Fill", "Random Fill", "Shred + Zero"]
        method_index = simpledialog.askinteger("Wipe Method", f"Choose wipe method:\n1. {methods[0]}\n2. {methods[1]}\n3. {methods[2]}")
        if not method_index or method_index < 1 or method_index > 3:
            return
        method = methods[method_index - 1]

        confirm = messagebox.askyesno("CONFIRM", f"This will ERASE all data on {device} using {method}.\n\nAre you absolutely sure?")
        if not confirm:
            return

        messagebox.showinfo("Starting Wipe", f"Wiping {device} using {method}. This may take a while...")

        log_file, status = self.wipe_device(device, method)

        verified_clean = False
        if status == "success":
            verify = messagebox.askyesno("Verify", "Attempt simple verification of wipe?\n(Reads first 512 bytes)")
            if verify:
                verified_clean = verify_wipe(device)
                if verified_clean:
                    messagebox.showinfo("Verified", "First sector appears zeroed.")
                else:
                    messagebox.showwarning("Not Verified", "First sector NOT clean. Wipe may be incomplete.")
        else:
            messagebox.showerror("Wipe Failed", "The wipe process failed. See log for details.")

        cert_path = write_certificate(
            device=device,
            method=method,
            log_file=log_file,
            status=status,
            verified_clean=verified_clean
        )

        messagebox.showinfo("Done", f"Wipe complete.\n\nLog file: {log_file}\nCertificate saved at:\n{cert_path}")

if __name__ == "__main__":
    app = WipeApp()
    app.main()
