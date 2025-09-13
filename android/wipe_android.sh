#!/bin/bash

set -e

# check for required tools or disply error message
for tool in dialog adb fastboot; do
    if ! command -v $tool &>/dev/null; then
        echo "Error: $tool is not installed or not in PATH."
        exit 1
    fi
done

# start adb and check for device
adb start-server
sleep 1

DEVICE_ID=$(adb get-serialno)

if [[ "$DEVICE_ID" == "unknown" || -z "$DEVICE_ID" ]]; then
    dialog --msgbox "No Android device detected.\n\nPlease connect a device with USB debugging enabled and authorize it." 8 60
    exit 1
fi

# get device informatiob (useful metadata)
DEVICE_MODEL=$(adb shell getprop ro.product.model | tr -d '\r')
DEVICE_MANUFACTURER=$(adb shell getprop ro.product.manufacturer | tr -d '\r')
LOCK_STATE=$(adb shell getprop ro.boot.verifiedbootstate | tr -d '\r')

# check for the bootloader's lock state
if [[ "$LOCK_STATE" == "green" ]]; then
    dialog --msgbox "Device: $DEVICE_MANUFACTURER $DEVICE_MODEL\n\nBootloader is LOCKED.\n\nThis device CANNOT be securely wiped using fastboot.\n\nPlease unlock the bootloader or use a supported device." 12 60
    clear
    exit 1
fi

# Confirm wipe
dialog --yesno "Device detected:\n$DEVICE_MANUFACTURER $DEVICE_MODEL\n\nAre you sure you want to WIPE ALL user data?" 10 60
response=$?

if [[ $response -ne 0 ]]; then
    dialog --msgbox "Operation cancelled." 6 30
    exit 0
fi

dialog --infobox "Rebooting device to fastboot mode..." 5 40
adb reboot bootloader

# wait up to 5 minutes for fastboot mode
MAX_WAIT=300
WAIT_TIME=0

dialog --infobox "Waiting for device to enter fastboot mode..." 5 50

while true; do
    FASTBOOT_ID=$(fastboot devices | awk '{print $1}')
    if [[ -n "$FASTBOOT_ID" ]]; then
        break
    fi
    sleep 1
    ((WAIT_TIME++))
    if (( WAIT_TIME >= MAX_WAIT )); then
        dialog --msgbox "Timed out waiting for device to enter fastboot mode.\n\nPlease check your connection and try again." 10 60
        exit 1
    fi
done

# Perform the wipe
dialog --infobox "Wiping userdata..." 5 40
fastboot -s "$FASTBOOT_ID" erase userdata

dialog --infobox "Wiping cache..." 5 40
fastboot -s "$FASTBOOT_ID" erase cache || true

# reboot confirmation
dialog --yesno "Wipe complete.\n\nDo you want to reboot the device now?" 8 50
reboot_response=$?

if [[ $reboot_response -eq 0 ]]; then
    fastboot -s "$FASTBOOT_ID" reboot
else
    dialog --msgbox "device left in fastboot mode.\nYou can manually reboot with:\nfastboot reboot" 10 60
fi

clear
