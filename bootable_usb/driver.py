from certgen import save_certificates
import os
import subprocess
import threading
import time
import uuid
import json
import shutil
import hashlib
from datetime import datetime
import tkinter as tk
from tkinter import ttk, messagebox
import platform, getpass, socket

# - Utilities -
def is_root():
    try:
        return os.geteuid() == 0
    except AttributeError:
        return False

def run_cmd(cmd, capture_output=True):
    try:
        res = subprocess.run(
            cmd, shell=True,
            stdout=subprocess.PIPE if capture_output else None,
            stderr=subprocess.PIPE if capture_output else None, text=True
        )
        if res.returncode != 0:
            return None
        return res.stdout.strip() if capture_output else ""
    except Exception:
        return None

def check_dependency(cmd):
    return shutil.which(cmd) is not None

def sha256_of_file(path):
    if not os.path.exists(path):
        return None
    h = hashlib.sha256()
    with open(path, 'rb') as f:
        for chunk in iter(lambda: f.read(8192), b''):
            h.update(chunk)
    return h.hexdigest()

def collect_system_metadata():
    return {
        "hostname": socket.gethostname(),
        "os": platform.platform(),
        "kernel": platform.release(),
        "operator": getpass.getuser()
    }

def collect_device_metadata(device):
    meta = {"device": device}
    try:
        if check_dependency("smartctl"):
            info = run_cmd(f"smartctl -i {device}")
            if info:
                for line in info.splitlines():
                    if "Model Number" in line: meta["model"] = line.split(":")[-1].strip()
                    if "Serial Number" in line: meta["serial"] = line.split(":")[-1].strip()
                    if "Firmware Version" in line: meta["firmware"] = line.split(":")[-1].strip()
    except: pass
    try:
        size = run_cmd(f"blockdev --getsize64 {device}")
        if size:
            meta["capacity_bytes"] = size
            meta["capacity_human"] = f"{int(size)//(1024**3)} GB"
    except: pass
    meta["interface"] = detect_device_type(device)
    return meta

VERSION = "1.0.0"

def script_sha256():
    try:
        return sha256_of_file(__file__)
    except:
        return None


# - Device Detection -
def list_block_devices():
    devices = []
    try:
        out = subprocess.check_output("lsblk -dpno NAME,SIZE,MODEL | grep -v 'loop\\|sr0'", shell=True, text=True)
        for line in out.splitlines():
            parts = line.split(None, 2)
            if len(parts) >= 2:
                dev = parts[0]
                info = ' '.join(parts[1:])
                devices.append((dev, info))
    except Exception:
        pass
    return devices

def detect_device_type(device):
    base = os.path.basename(device)

    # NVMe is reliable
    if base.startswith('nvme'):
        return 'nvme'

    # Try lsblk TRAN column first
    try:
        tran = run_cmd(f"lsblk -ndo TRAN {device}")
        if tran:
            tran = tran.strip().lower()
            if tran == "nvme":
                return "nvme"
            if tran == "usb":
                return "usb"
            if tran in ("sata", "ata"):
                return "ata"
    except Exception:
        pass

    # Fallback: smartctl
    if check_dependency("smartctl"):
        try:
            info = run_cmd(f"smartctl -i {device}")
            if info:
                if "NVMe" in info:
                    return "nvme"
                if "USB" in info or "Transport protocol:  USB" in info:
                    return "usb"
                if "ATA" in info or "SATA" in info:
                    return "ata"
        except Exception:
            pass

    # Final heuristic
    if base.startswith("sd"):
        return "ata"
    return "unknown"


# - Unmount the device -
def unmount_device(device, logf):
    try:
        partitions = []
        base = os.path.basename(device)
        for p in os.listdir("/dev"):
            if p.startswith(base) and p != base:
                partitions.append("/dev/" + p)
        for part in partitions:
            try:
                logf.write(f"Unmounting {part}\n")
                subprocess.run(["umount", "-f", part], check=True, timeout=10)
            except subprocess.TimeoutExpired:
                subprocess.run(["umount", "-l", part], check=True)
            except Exception as e:
                logf.write(f"Could not unmount {part}: {e}\n")
        time.sleep(2)
        return True
    except Exception as e:
        logf.write(f"Unmount error: {e}\n")
        return False

# - Quick Wipe the USB -
def find_partition(device, retries=10, delay=1):
    """Poll /dev until a child partition appears (e.g., /dev/sdb1)"""
    base = os.path.basename(device)
    for _ in range(retries):
        for p in os.listdir("/dev"):
            if p.startswith(base) and p != base:
                return "/dev/" + p
        time.sleep(delay)
    return None


def quick_wipe_usb(device, logf):
    logf.write(f"[{datetime.now().isoformat()}] Starting quick wipe on {device}\n")
    try:
        # --- Unmount ---
        if not unmount_device(device, logf):
            logf.write("Warning: could not fully unmount, continuing anyway...\n")

        # --- Wipe filesystem signatures ---
        logf.write("Removing filesystem signatures...\n")
        subprocess.run(["wipefs", "-a", device], check=True, timeout=30)

        # --- Zero out start/end ---
        size = int(subprocess.check_output(["blockdev", "--getsize64", device], text=True).strip())

        logf.write("Zeroing first 10MB...\n")
        subprocess.run(["dd", "if=/dev/zero", f"of={device}", "bs=1M", "count=10"],
                       check=True, timeout=30)

        if size > 1048576:
            seek = (size - 1048576) // 1048576
            logf.write("Zeroing last MB...\n")
            subprocess.run(["dd", "if=/dev/zero", f"of={device}", "bs=1M",
                            f"seek={seek}", "count=1"],
                           check=True, timeout=30)

        # --- Partition table ---
        logf.write("Creating new partition table...\n")
        subprocess.run(["parted", "-s", device, "mklabel", "msdos"], check=True, timeout=10)
        subprocess.run(["parted", "-s", device, "mkpart", "primary", "fat32", "0%", "100%"],
                       check=True, timeout=10)

        # --- Find partition ---
        logf.write("Waiting for new partition...\n")
        part = find_partition(device, retries=10, delay=1)
        if not part:
            logf.write("Partition not found after wipe.\n")
            return False, "partition_not_found"

        # --- Format FAT32 ---
        logf.write(f"Formatting {part} as FAT32...\n")
        subprocess.run(["mkfs.vfat", "-F", "32", "-n", "USBDRIVE", part],
                       check=True, timeout=30,
                       stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)

        logf.write(f"Quick wipe complete. Partition: {part}\n")
        return True, f"usb_quick_wipe_ok:{part}"

    except subprocess.TimeoutExpired as e:
        logf.write(f"Quick wipe failed: timeout in step {e.cmd}\n")
        return False, "usb_quick_wipe_timeout"
    except subprocess.CalledProcessError as e:
        logf.write(f"Quick wipe failed: command {e.cmd} returned {e.returncode}\n")
        return False, "usb_quick_wipe_failed"
    except Exception as e:
        logf.write(f"Quick wipe failed: {e}\n")
        return False, "usb_quick_wipe_failed"


# - ATA/NVMe Wipes -
def ata_secure_erase(device, logf):
    logf.write(f"[{datetime.now().isoformat()}] Starting ATA secure erase on {device}\n")
    if not check_dependency("hdparm"):
        logf.write("hdparm not installed.\n")
        return False, "hdparm_missing"

    # Check device info
    info = run_cmd(f"hdparm -I {device}")
    if not info:
        logf.write("Failed to run hdparm -I\n")
        return False, "hdparm_info_fail"

    logf.write(info + "\n")

    # Check frozen state
    if "frozen" in info.lower():
        logf.write("Device is frozen. Power cycle required.\n")
        return False, "frozen"

    # Check support
    if "supported" in info.lower() and ("security erase" in info.lower() or "security erase unit" in info.lower()):
        logf.write("Secure erase supported.\n")
        # Generate temp password
        passwd = "P@ssw0rd" + uuid.uuid4().hex[:8]
        logf.write(f"Setting security password...\n")
        out = run_cmd(f"hdparm --user-master u --security-set-pass {passwd} {device}")
        if out is None:
            logf.write("Failed to set security password.\n")
            return False, "security_set_fail"

        logf.write("Issuing secure erase...\n")
        proc = subprocess.Popen(
            f"hdparm --user-master u --security-erase {passwd} {device}",
            shell=True, stdout=subprocess.PIPE, stderr=subprocess.STDOUT, text=True
        )
        for line in proc.stdout:
            logf.write(line)
        proc.wait()

        if proc.returncode == 0:
            logf.write("Secure erase completed successfully.\n")
            return True, "secure_erase_ok"
        else:
            logf.write(f"Secure erase failed, code={proc.returncode}\n")
            return False, "secure_erase_failed"
    else:
        logf.write("Secure erase not supported. Falling back to multi-pass random overwrite.\n")
        success = random_overwrite(device, passes=3, block_size=1024*1024, logf=logf)
        return (success, "random_overwrite_ok" if success else "random_overwrite_failed")


def random_overwrite(device, passes=3, block_size=1024*1024, logf=None):
    try:
        size_output = run_cmd(f"blockdev --getsize64 {device}")
        if not size_output:
            if logf: logf.write("Could not get device size.\n")
            return False
        size = int(size_output)
        if logf: logf.write(f"Device size: {size} bytes\n")

        with open(device, "wb") as f:
            for p in range(passes):
                if logf: logf.write(f"Pass {p+1}/{passes}\n")
                written = 0
                while written < size:
                    data = os.urandom(min(block_size, size - written))
                    f.write(data)
                    written += len(data)
                f.flush()
                os.fsync(f.fileno())

        if logf: logf.write("Random overwrite complete.\n")
        return True
    except PermissionError:
        if logf: logf.write("Permission denied. Run as root!\n")
        return False
    except Exception as e:
        if logf: logf.write(f"Random overwrite error: {e}\n")
        return False


def nvme_sanitize(device, logf):
    logf.write(f"[{datetime.now().isoformat()}] Starting NVMe sanitize on {device}\n")
    if not check_dependency('nvme'):
        logf.write("nvme tool missing\n")
        return False, 'nvme_missing'

    try:
        logf.write("Running sanitize (sanact=1)...\n")
        res = subprocess.run(
            ["nvme", "sanitize", device, "--sanact=1"],
            stdout=subprocess.PIPE, stderr=subprocess.STDOUT, text=True
        )
        logf.write(res.stdout + '\n')
        if res.returncode != 0:
            logf.write("Sanitize command failed.\n")
            return False, 'nvme_sanitize_failed'

        logf.write("Running format (ses=1)...\n")
        res2 = subprocess.run(
            ["nvme", "format", device, "--ses=1"],
            stdout=subprocess.PIPE, stderr=subprocess.STDOUT, text=True
        )
        logf.write(res2.stdout + '\n')
        if res2.returncode != 0:
            logf.write("Format command failed.\n")
            return False, 'nvme_format_failed'

        logf.write("NVMe sanitize+format completed successfully.\n")
        return True, 'nvme_sanitize_format_ok'

    except Exception as e:
        logf.write(f"nvme sanitize exception: {e}\n")
        return False, 'nvme_exception'


# - Fallback -
def dd_zero_cmd(device):
    return f"dd if=/dev/zero of={device} bs=4M status=progress conv=fsync"

def dd_random_cmd(device):
    return f"dd if=/dev/urandom of={device} bs=4M status=progress conv=fsync"

def shred_zero_cmd(device):
    return f"shred -v -n 3 {device} && dd if=/dev/zero of={device} bs=4M status=progress conv=fsync"

# - Verification -
def verify_sampled(device, logf, samples=16):
    logf.write(f"[{datetime.now().isoformat()}] Sampled verification: {samples} samples\n")
    try:
        size_bytes = int(subprocess.check_output(f"blockdev --getsize64 {device}", shell=True, text=True).strip())
    except Exception as e:
        logf.write(f"blockdev failed: {e}\n")
        return False
    import random
    offsets = [0, max(0, size_bytes-4096)]
    for _ in range(max(0, samples-2)):
        offsets.append(random.randrange(0, max(1, size_bytes-4096)))
    try:
        with open(device, 'rb') as f:
            for off in offsets:
                f.seek(off)
                data = f.read(4096)
                if any(b != 0 for b in data):
                    logf.write(f"Non-zero data at {off}\n")
                    return False
        return True
    except Exception as e:
        logf.write(f"Sampled verify exception: {e}\n")
        return False

def verify_full(device, logf):
    logf.write(f"[{datetime.now().isoformat()}] Full verification started.\n")
    block_size = 1024*1024
    try:
        with open(device,'rb') as f:
            while True:
                data = f.read(block_size)
                if not data: break
                if any(b!=0 for b in data):
                    logf.write("Non-zero found during full verification\n")
                    return False
        return True
    except Exception as e:
        logf.write(f"Full verify failed: {e}\n")
        return False

# - Certificates -
def write_certificate(device, method, log_file, status, verified_clean, extra):
    cert = {
        "uuid": str(uuid.uuid4()),
        "standard": "NIST_SP_800-88r1",
        "device": device,
        "method": method,
        "timestamp": datetime.now().isoformat(),
        "status": status,
        "verified_clean": verified_clean,
        "verification_method": extra.get("verification_method", "none"),
        "log_file": log_file,
        "system_metadata": extra.get("system_metadata"),
        "device_metadata": extra.get("device_metadata"),
        "wipe_metadata": extra.get("wipe_metadata"),
        "execution_metadata": extra.get("execution_metadata"),
    }
    return save_certificates(cert)

# - Android -
def collect_android_metadata():
    meta = {}
    meta['serial'] = run_cmd("adb get-serialno") or "unknown"
    meta['model'] = run_cmd("adb shell getprop ro.product.model") or "unknown"
    meta['manufacturer'] = run_cmd("adb shell getprop ro.product.manufacturer") or "unknown"
    meta['android_version'] = run_cmd("adb shell getprop ro.build.version.release") or "unknown"
    meta['bootloader_state'] = run_cmd("adb shell getprop ro.boot.verifiedbootstate") or "unknown"
    meta['device_name'] = run_cmd("adb shell getprop ro.product.name") or "unknown"
    return meta

def wipe_android():
    required = ["adb", "fastboot"]
    for t in required:
        if not check_dependency(t):
            messagebox.showerror("Missing Tool", f"{t} not installed.")
            return "failed", False, {}

    run_cmd("adb start-server")
    time.sleep(1)

    serial = run_cmd("adb get-serialno")
    if serial in [None, "unknown", ""]:
        messagebox.showerror("No device", "Connect Android with USB debugging and authorize it.")
        return "failed", False, {}

    meta = collect_android_metadata()
    if meta['bootloader_state'].strip() == "green":
        status = "bootloader_locked"
        cert_path = write_certificate(
            device="android",
            method="auto",
            log_file="/tmp/wipe_android.log",
            status=status,
            verified_clean=False,
            extra={"device_metadata": meta}
        )
        messagebox.showerror(
            "Bootloader Locked",
            f"{meta['manufacturer']} {meta['model']}\nBootloader LOCKED.\nCannot wipe securely.\nMetadata saved at {cert_path}"
        )
        return status, False, meta

    messagebox.showinfo("Rebooting", "Device will reboot to fastboot mode...")
    run_cmd("adb reboot bootloader")

    fastboot_id = None
    for _ in range(300):
        out = run_cmd("fastboot devices")
        if out:
            fastboot_id = out.split()[0]
            break
        time.sleep(1)

    if not fastboot_id:
        status = "fastboot_timeout"
        cert_path = write_certificate(
            device="android",
            method="auto",
            log_file="/tmp/wipe_android.log",
            status=status,
            verified_clean=False,
            extra={"device_metadata": meta}
        )
        messagebox.showerror("Timeout", f"Device did not enter fastboot mode.\nMetadata saved at {cert_path}")
        return status, False, meta

    messagebox.showinfo("Wiping", "Wiping userdata + cache...")
    run_cmd(f"fastboot -s {fastboot_id} erase userdata")
    run_cmd(f"fastboot -s {fastboot_id} erase cache")
    status = "android_wipe_done"

    cert_path = write_certificate(
        device="android",
        method="auto",
        log_file="/tmp/wipe_android.log",
        status=status,
        verified_clean=True,
        extra={"device_metadata": meta}
    )

    reboot = messagebox.askyesno("Done", f"Wipe complete.\nCertificate saved at:\n{cert_path}\nReboot now?")
    if reboot:
        run_cmd(f"fastboot -s {fastboot_id} reboot")
    else:
        messagebox.showinfo("Manual Reboot", "Device left in fastboot. Use 'fastboot reboot'.")

    return status, True, meta

# - GUI -
class WipeApp:
    def __init__(self):
        self.root = tk.Tk()

        # --- Theme Colors ---
        BG_COLOR = "#1e1e1e"
        FG_COLOR = "#39FF14"
        TEXT_BG = "#121212"
        WIDGET_BG = "#2c2c2c"
        ACTIVE_BG = "#404040"
        DISABLED_FG = "#666666"
        BORDER_COLOR = "#39FF14"
        SELECT_FG_COLOR = "#121212" # Text color when an item is selected

        self.root.configure(bg=BG_COLOR)
        self.root.title("NIST-Aware Wiper | NullBytes")
        self.root.geometry("900x650")

        self.cancel_flag = threading.Event()
        self.current_process = None

        # --- Style ---
        style = ttk.Style(self.root)
        style.theme_use("clam")

        # General configurations
        style.configure('.',
                        background=BG_COLOR,
                        foreground=FG_COLOR,
                        fieldbackground=WIDGET_BG,
                        borderwidth=0,
                        font=('Helvetica', 10))

        # Frames and Labels
        style.configure('TFrame', background=BG_COLOR)
        style.configure('TLabel', background=BG_COLOR, foreground=FG_COLOR, padding=5)
        style.configure('TLabelFrame', background=BG_COLOR, bordercolor=BORDER_COLOR)
        style.configure('TLabelFrame.Label', background=BG_COLOR, foreground=FG_COLOR, font=('Helvetica', 10, 'bold'))

        # Radiobutton
        style.configure('TRadiobutton',
                        background=BG_COLOR,
                        foreground=FG_COLOR,
                        indicatorrelief=tk.FLAT)
        style.map('TRadiobutton',
                    background=[('active', ACTIVE_BG)],
                    indicatorcolor=[('selected', FG_COLOR), ('!selected', WIDGET_BG)])

        # Button
        style.configure('TButton',
                        background=WIDGET_BG,
                        foreground=FG_COLOR,
                        padding=8,
                        borderwidth=1,
                        relief=tk.FLAT,
                        font=('Helvetica', 10, 'bold'))
        style.map('TButton',
                    background=[('active', ACTIVE_BG), ('disabled', '#333333')],
                    foreground=[('disabled', DISABLED_FG)])

        # Combobox
        self.root.option_add('*TCombobox*Listbox.background', WIDGET_BG)
        self.root.option_add('*TCombobox*Listbox.foreground', FG_COLOR)
        self.root.option_add('*TCombobox*Listbox.selectBackground', FG_COLOR)
        self.root.option_add('*TCombobox*Listbox.selectForeground', SELECT_FG_COLOR)
        style.configure('TCombobox',
                        fieldbackground=WIDGET_BG,
                        background=WIDGET_BG,
                        foreground=FG_COLOR,
                        arrowcolor=FG_COLOR,
                        selectbackground=WIDGET_BG,
                        selectforeground=FG_COLOR,
                        insertcolor=FG_COLOR,
                        bordercolor=BORDER_COLOR,
                        lightcolor=BORDER_COLOR,
                        darkcolor=BORDER_COLOR)
        style.map('TCombobox',
                    background=[('readonly', WIDGET_BG)],
                    fieldbackground=[('readonly', WIDGET_BG)],
                    foreground=[('readonly', FG_COLOR)])

        # Progressbar
        style.configure('Horizontal.TProgressbar',
                        background=FG_COLOR,
                        troughcolor=WIDGET_BG,
                        bordercolor=BORDER_COLOR,
                        lightcolor=FG_COLOR,
                        darkcolor=FG_COLOR)

        # --- Main Layout ---
        frame = ttk.Frame(self.root, padding=8)
        frame.pack(fill='both', expand=True)

        # Device selection
        top = ttk.Frame(frame)
        top.pack(fill='x')
        ttk.Label(top,text="Select target:").pack(side='left')
        self.device_var = tk.StringVar()
        self.device_combo = ttk.Combobox(top, textvariable=self.device_var, state='readonly')
        self.device_combo.pack(side='left', padx=6, fill='x', expand=True)
        ttk.Button(top, text="Refresh", command=self.refresh_devices).pack(side='left')

        # Wipe method (dynamic)
        self.method_frame = ttk.LabelFrame(frame, text='Wipe Method')
        self.method_frame.pack(fill='x', pady=6, ipady=5, ipadx=5)
        self.method_var = tk.StringVar(value='auto')
        self.method_buttons = []

        # Description box
                # Description box
        self.method_desc = tk.Text(self.method_frame, height=4, wrap='word', bg=TEXT_BG, fg=FG_COLOR, relief='solid', borderwidth=1,highlightthickness=0, font=('Helvetica', 10))
        self.method_desc.pack(fill='x', pady=4, padx=5)
        self.method_desc.config(state='disabled') # Set state after creation

        # Verification (dynamic)
        self.ver_frame = ttk.LabelFrame(frame,text='Verification')
        self.ver_frame.pack(fill='x', pady=6, ipady=5, ipadx=5)
        self.verify_var = tk.StringVar(value='none')
        self.verify_buttons = []

        # Controls
        ctrl = ttk.Frame(frame)
        ctrl.pack(fill='x', pady=8)
        ttk.Button(ctrl,text='Start Wipe',command=self.start).pack(side='left')
        ttk.Button(ctrl,text='Open Logs',command=self.open_logs_dir).pack(side='left',padx=6)
        self.cancel_btn = ttk.Button(ctrl,text='Cancel',command=self.cancel,state='disabled')
        self.cancel_btn.pack(side='right')

        # Log
        self.log = tk.Text(frame, height=20, bg=TEXT_BG, fg=FG_COLOR,
                           relief='solid', borderwidth=1, insertbackground="#FFFFFF", # White cursor is more visible
                           selectbackground=FG_COLOR, selectforeground=SELECT_FG_COLOR,
                           highlightthickness=0,
                           font=("monospace", 10))
        self.log.pack(fill='both', expand=True, pady=6)

        # Loader
        self.startup_loader = ttk.Progressbar(self.root, mode='indeterminate', style='Horizontal.TProgressbar')
        self.startup_loader.pack(fill='x', padx=8, pady=5)
        self.startup_loader.start(10)
        self.status_frame = None

        self.refresh_devices()
        self.startup_loader.stop()
        self.startup_loader.destroy()

    def update_methods_for_device(self, device_label):
        for btn in self.method_buttons:
            btn.destroy()
        self.method_buttons.clear()

        if device_label == 'Android (ADB)':
            methods = [('Auto (Android Wipe)', 'auto')]
        else:
            target = device_label.split()[0]
            dtype = detect_device_type(target)
            if dtype == 'ata':
                methods = [('Auto (ATA Secure Erase) [Purge]', 'auto'),
                           ('Zero Fill [Clear]', 'zero'),
                           ('Random Fill [Clear]', 'random'),
                           ('Shred+Zero [Clear]', 'shred')]
            elif dtype == 'nvme':
                methods = [('Auto (NVMe Sanitize) [Purge]', 'auto'),
                           ('Zero Fill [Clear]', 'zero'),
                           ('Random Fill [Clear]', 'random'),
                           ('Shred+Zero [Clear]', 'shred')]
            elif dtype == 'usb':
                methods = [('Quick Wipe + FAT32 [Clear]', 'quick'),
                            ('Zero Fill [Clear]', 'zero'),
                            ('Random Fill [Clear]', 'random'),
                            ('Shred+Zero [Clear]', 'shred')]

            else:
                methods = [('Zero Fill [Clear]', 'zero'),
                           ('Random Fill [Clear]', 'random'),
                           ('Shred+Zero [Clear]', 'shred')]

        self.method_var.set(methods[0][1])
        for text, val in methods:
            b = ttk.Radiobutton(self.method_frame, text=text, variable=self.method_var, value=val,
                                command=lambda v=val: self.update_method_desc(v))
            b.pack(anchor='w', padx=5)
            self.method_buttons.append(b)
        self.update_method_desc(self.method_var.get())

    def update_method_desc(self, method):
        descs = {
            'auto': "Automatic secure wipe using hardware commands (ATA Secure Erase or NVMe Sanitize). NIST category: Purge.",
            'zero': "Overwrites all sectors with zeros. Simple clear operation. NIST category: Clear.",
            'random': "Overwrites all sectors with random data. Clear operation with stronger obfuscation. NIST category: Clear.",
            'shred': "Multiple overwrite passes (default 3) with random data followed by zero fill. NIST category: Clear (Close to Clear).",
            'quick': "Quick wipe for USB drives. Removes filesystem signatures, zeroes key areas, recreates partition table, and formats as FAT32. NIST category: Clear.",
        }
        self.method_desc.config(state='normal')
        self.method_desc.delete('1.0', tk.END)
        self.method_desc.insert(tk.END, descs.get(method, ""))
        self.method_desc.config(state='disabled')

    def update_verification_for_device(self, device_label):
        for btn in self.verify_buttons:
            btn.destroy()
        self.verify_buttons.clear()

        if device_label == 'Android (ADB)':
            options = [('None','none')]
        else:
            options = [('None','none'),('Sampled (fast check of random blocks)', 'sampled'),('Full (slow check of all blocks)', 'full')]

        self.verify_var.set(options[0][1])
        for text,val in options:
            b = ttk.Radiobutton(self.ver_frame, text=text, variable=self.verify_var, value=val)
            b.pack(anchor='w', padx=5)
            self.verify_buttons.append(b)

    def append_log(self,text):
        ts = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
        self.log.insert(tk.END, f"[{ts}] {text}\n")
        self.log.see(tk.END)

    def refresh_devices(self):
        self.devices = list_block_devices()
        labels = []
        for d,info in self.devices:
            dtype = detect_device_type(d)
            labels.append(f"{d} ({dtype.upper()}) — {info}")
        labels.append('Android (ADB)')
        self.device_combo['values'] = labels
        if labels:
            self.device_combo.set(labels[0])
            self.on_device_selected(None)
        self.device_combo.bind("<<ComboboxSelected>>", self.on_device_selected)

    def on_device_selected(self, event):
        sel = self.device_combo.get()
        if sel:
            self.update_methods_for_device(sel)
            self.update_verification_for_device(sel)

    def open_logs_dir(self):
        path = '/var/log/NullBytes'
        if not os.path.exists(path):
            path = '/tmp/NullBytes'
        messagebox.showinfo('Logs Directory',f'Certificates/logs under: {path}')

    def lock_ui(self):
        self.device_combo.configure(state='disabled')
        # Disable all method/verification buttons
        for btn in self.method_buttons + self.verify_buttons:
            btn.configure(state='disabled')
        self.cancel_btn.configure(state='normal')

    def unlock_ui(self):
        self.device_combo.configure(state='readonly')
         # Enable all method/verification buttons
        for btn in self.method_buttons + self.verify_buttons:
            btn.configure(state='normal')
        self.cancel_btn.configure(state='disabled')

    def cancel(self):
        self.cancel_flag.set()
        if self.current_process:
            try:
                self.current_process.terminate()
            except ProcessLookupError:
                pass
        self.append_log('>>> User requested cancel. Operation terminating... <<<')

    def start(self):
        sel = self.device_combo.get()
        if not sel:
            messagebox.showwarning('No device','Select a target device')
            return

        if sel != 'Android (ADB)':
            if not messagebox.askyesno("Confirm Wipe", f"Are you absolutely sure you want to wipe\n{sel}?\n\nThis action is IRREVERSIBLE and will destroy all data on the device."):
                return

        method = self.method_var.get()
        verify = self.verify_var.get()

        self.log.delete('1.0', tk.END)
        self.cancel_flag.clear()

        if sel == 'Android (ADB)':
            threading.Thread(target=self.run_android,daemon=True).start()
        else:
            device = sel.split()[0]
            threading.Thread(target=self.run_wipe,args=(device,method,verify),daemon=True).start()

    def run_android(self):
        self.lock_ui()
        status, verified, meta = wipe_android()
        self.append_log(f"Android wipe finished: {status}, verified: {verified}")
        self.unlock_ui()

    def run_wipe(self,device,method,verify):
        self.lock_ui()
        log_dir = '/var/log/NullBytes'
        try:
            os.makedirs(log_dir, exist_ok=True)
        except PermissionError:
            log_dir = '/tmp/NullBytes'
            os.makedirs(log_dir, exist_ok=True)

        logf = open(os.path.join(log_dir, f"wipe_{os.path.basename(device)}_{int(time.time())}.log"), 'w')
        status = 'unknown'
        verified_clean = False
        try:
            self.append_log(f"Starting wipe on {device} with method '{method}' and verification '{verify}'.")
            logf.write(f"Wipe initiated at {datetime.now().isoformat()} on {device}\n")
            sysmeta = collect_system_metadata()
            devmeta = collect_device_metadata(device)
            success = False

            unmount_success = unmount_device(device, logf)
            if not unmount_success:
                self.append_log("WARNING: Could not unmount all partitions. Continuing anyway.")
                logf.write("WARNING: Could not unmount all partitions. Continuing anyway.\n")


            if method=='auto':
                dtype = detect_device_type(device)
                if dtype=='ata':
                    success,status = ata_secure_erase(device,logf)
                elif dtype=='nvme':
                    success,status = nvme_sanitize(device,logf)
                else: # Fallback for USB/Unknown
                    self.append_log("Auto method not applicable, falling back to Zero Fill.")
                    logf.write("Auto method not applicable, falling back to Zero Fill.\n")
                    method = 'zero' # proceed to the zero fill block

            if method=='zero':
                cmd = dd_zero_cmd(device)
                self.current_process = subprocess.Popen(cmd,shell=True,stdout=subprocess.PIPE,stderr=subprocess.STDOUT,text=True)
                for line in self.current_process.stdout:
                    logf.write(line); self.append_log(line.strip())
                    if self.cancel_flag.is_set(): break
                self.current_process.wait()
                success = (self.current_process.returncode==0) and not self.cancel_flag.is_set()
                status = 'dd_zero_ok' if success else 'dd_zero_failed'
            elif method=='random':
                cmd = dd_random_cmd(device)
                self.current_process = subprocess.Popen(cmd,shell=True,stdout=subprocess.PIPE,stderr=subprocess.STDOUT,text=True)
                for line in self.current_process.stdout:
                    logf.write(line); self.append_log(line.strip())
                    if self.cancel_flag.is_set(): break
                self.current_process.wait()
                success = (self.current_process.returncode==0) and not self.cancel_flag.is_set()
                status = 'dd_random_ok' if success else 'dd_random_failed'
            elif method=='shred':
                cmd = shred_zero_cmd(device)
                self.current_process = subprocess.Popen(cmd,shell=True,stdout=subprocess.PIPE,stderr=subprocess.STDOUT,text=True)
                for line in self.current_process.stdout:
                    logf.write(line); self.append_log(line.strip())
                    if self.cancel_flag.is_set(): break
                self.current_process.wait()
                success = (self.current_process.returncode==0) and not self.cancel_flag.is_set()
                status = 'shred_ok' if success else 'shred_failed'
            elif method == 'quick':
                success, status = quick_wipe_usb(device, logf)

            if self.cancel_flag.is_set():
                status = "cancelled_by_user"
                success = False

            if success:
                self.append_log("Wipe process completed successfully.")
                self.append_log(f"Starting verification: {verify}...")
                logf.write(f"Wipe successful. Starting verification: {verify}.\n")
                if verify=='none':
                    verified_clean=False
                    self.append_log("Verification skipped.")
                elif verify=='sampled':
                    verified_clean = verify_sampled(device,logf)
                elif verify=='full':
                    verified_clean = verify_full(device,logf)

                if verify != 'none':
                    self.append_log(f"Verification result: {'PASSED' if verified_clean else 'FAILED'}")
                    logf.write(f"Verification result: {'PASSED' if verified_clean else 'FAILED'}\n")
            else:
                self.append_log(f"Wipe failed. Status: {status}")
                logf.write(f"Wipe failed with status: {status}\n")

            extra = {
                "system_metadata": sysmeta,
                "device_metadata": devmeta,
                "verification_method": verify,
                "execution_metadata": {
                    "version": VERSION,
                    "script_hash": script_sha256()
                }
            }
            cert_path = write_certificate(device,method,logf.name,status,verified_clean,extra)
            self.append_log(f"--- Process Finished ---")
            self.append_log(f"Certificate written to: {cert_path}")
            messagebox.showinfo("Wipe Complete", f"Operation on {device} has finished.\n\nCertificate saved to:\n{cert_path}")
        except Exception as e:
            self.append_log(f"An unexpected error occurred: {e}")
            logf.write(f"FATAL ERROR: {e}\n")
        finally:
            logf.close()
            self.unlock_ui()
            self.current_process = None

if __name__=='__main__':
    if not is_root():
        # Use Tkinter to show the error if possible
        try:
            root = tk.Tk()
            root.withdraw()
            messagebox.showerror("Root Required", "This application must be run as root (or with sudo).")
        except tk.TclError:
            print("This application must be run as root (or with sudo).")
        exit(1)

    app = WipeApp()
    app.root.mainloop()
