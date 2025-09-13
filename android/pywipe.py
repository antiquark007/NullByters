import subprocess
import tkinter as tk
from tkinter import messagebox
import time
import sys


def run_cmd(command, capture_output=True):
    try:
        result = subprocess.run(command, shell=True, capture_output=capture_output, text=True)
        if result.returncode != 0:
            return None
        return result.stdout.strip()
    except Exception as e:
        return None

def check_dependency(cmd):
    return run_cmd(f"which {cmd}") is not None

def show_error(msg):
    messagebox.showerror("Error", msg)
    sys.exit(1)


required_tools = ["adb", "fastboot"]
for tool in required_tools:
    if not check_dependency(tool):
        show_error(f"'{tool}' is not installed or not in PATH.")


root = tk.Tk()
root.withdraw()

# starts the server and detects android devices
run_cmd("adb start-server")
time.sleep(1)

device_id = run_cmd("adb get-serialno")
if device_id in ["unknown", None, ""]:
    show_error("No Android device detected.\nPlease connect a device with USB debugging enabled and authorize it.")

# gets the useful metadata
device_model = run_cmd("adb shell getprop ro.product.model") or "Unknown"
device_manu = run_cmd("adb shell getprop ro.product.manufacturer") or "Unknown"
lock_state = run_cmd("adb shell getprop ro.boot.verifiedbootstate") or "unknown"

# if the bootloader is locked shows green and gives the user an output before exiting
if lock_state.strip() == "green":
    show_error(
        f"Device: {device_manu} {device_model}\n\n"
        "Bootloader is LOCKED.\n"
        "This device CANNOT be securely wiped using fastboot.\n\n"
        "Please unlock the bootloader (if possible) or use a supported device."
    )

# wipes out everything
proceed = messagebox.askyesno("Confirm Wipe", f"Device detected:\n{device_manu} {device_model}\n\nProceed to WIPE ALL user data?")
if not proceed:
    messagebox.showinfo("Cancelled", "Operation cancelled.")
    sys.exit(0)


messagebox.showinfo("Rebooting", "Device will reboot to fastboot mode...")
run_cmd("adb reboot bootloader")


messagebox.showinfo("Please Wait", "Waiting for device to enter fastboot mode...\n(Max 5 minutes)")

max_wait = 300
elapsed = 0
fastboot_id = None

# have the app wait for 5 mins for the mobile to load fastboot mode
while elapsed < max_wait:
    output = run_cmd("fastboot devices")
    if output:
        fastboot_id = output.split()[0]
        break
    time.sleep(1)
    elapsed += 1

if not fastboot_id:
    show_error("Timed out waiting for device to enter fastboot mode.\nPlease check your connection.")


messagebox.showinfo("Wiping", "Wiping userdata partition...")
run_cmd(f"fastboot -s {fastboot_id} erase userdata")

messagebox.showinfo("Wiping", "Wiping cache partition...")
run_cmd(f"fastboot -s {fastboot_id} erase cache")

# confirm reboot
reboot = messagebox.askyesno("Done", "Wipe complete.\n\nDo you want to reboot the device now?")
if reboot:
    run_cmd(f"fastboot -s {fastboot_id} reboot")
else:
    messagebox.showinfo("Manual Reboot", "Device left in fastboot mode.\nYou can manually reboot with:\nfastboot reboot")
