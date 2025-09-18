import subprocess
import time

def run_cmd(cmd, capture_output=True):
    try:
        result = subprocess.run(cmd, shell=True, capture_output=capture_output, text=True, timeout=15)
        if result.returncode == 0:
            return result.stdout.strip()
        else:
            return None
    except Exception as e:
        return None

def adb_devices():
    output = run_cmd("adb devices")
    if output:
        lines = output.splitlines()
        devices = [line.split()[0] for line in lines[1:] if "device" in line]
        return devices
    return []

def is_rooted():
    su_path = run_cmd("adb shell which su")
    return su_path is not None and su_path != ""

def bootloader_unlocked():
    prop = run_cmd("adb shell getprop ro.boot.flash.locked")
    if prop is not None:
        return prop == "0"
    return False

def wait_for_fastboot(timeout=30):
    print("Waiting for device to appear in fastboot mode...")
    for i in range(timeout):
        output = run_cmd("fastboot devices")
        if output and output.strip() != "":
            print(f"Fastboot device detected: {output.strip()}")
            return True
        time.sleep(1)
    print("Timeout waiting for fastboot device.")
    return False

def main():
    devices = adb_devices()
    if not devices:
        print("No device connected via ADB.")
        return

    device = devices[0]
    print(f"Device detected: {device}")

    if is_rooted():
        print("Device appears to be rooted.")
        print("Attempting to reboot into recovery...")
        result = run_cmd("adb shell su -c 'reboot recovery'")
        if result is None:
            print("Failed to reboot recovery via su.")
        else:
            print("Rebooted to recovery.")
        print("You may need to wipe manually in recovery if auto-wipe fails.")
        return

    if bootloader_unlocked():
        print("Bootloader is unlocked. Using Fastboot to wipe device.")
        print("Rebooting to bootloader...")
        run_cmd("adb reboot bootloader")

        if wait_for_fastboot():
            print("Erasing userdata and cache...")
            run_cmd("fastboot erase userdata")
            run_cmd("fastboot erase cache")
            run_cmd("fastboot reboot")
            print("Wipe complete. Device rebooting.")
        else:
            print("[!] Could not detect fastboot device. Abort.")
        return

    print("Bootloader is locked and device is not rooted.")
    print("Cannot wipe device. Your phone is protected by the manufacturer or carrier.")

if __name__ == "__main__":
    main()
