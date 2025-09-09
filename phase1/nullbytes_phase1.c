// nullbytes_phase1.c
// Phase 1: enumerate block devices (excluding system/boot disk) and write device_inventory.json
// Compile: gcc $(pkg-config --cflags --libs libudev json-c) nullbytes_phase1.c -o nullbytes_phase1
// Run: sudo ./nullbytes_phase1

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <libudev.h>
#include <json-c/json.h>

#define BUF_LEN 512

/* Return boot disk path (e.g. "/dev/sda") or NULL if not found */
char *get_boot_disk() {
    FILE *fp;
    char part[BUF_LEN] = {0};
    char pkname[BUF_LEN] = {0};
    char *diskpath = NULL;

    /* Ask findmnt for the parent device name (PKNAME) of root mount */
    fp = popen("findmnt -n -o PKNAME / 2>/dev/null", "r");
    if (!fp) return NULL;
    if (fgets(pkname, sizeof(pkname), fp) == NULL) {
        pclose(fp);
        return NULL;
    }
    pclose(fp);

    /* strip newline */
    pkname[strcspn(pkname, "\n")] = 0;
    if (strlen(pkname) == 0) return NULL;

    diskpath = malloc(BUF_LEN);
    if (!diskpath) return NULL;
    snprintf(diskpath, BUF_LEN, "/dev/%s", pkname);
    return diskpath;
}

/* safe getter for sysattr or property with fallback */
const char *get_attr_fallback(struct udev_device *dev, const char *first, const char *second) {
    const char *v = NULL;
    if (first) v = udev_device_get_sysattr_value(dev, first);
    if (!v && second) v = udev_device_get_sysattr_value(dev, second);
    if (!v) v = udev_device_get_property_value(dev, first ? first : second);
    return v;
}

int main(void) {
    struct udev *udev = NULL;
    struct udev_enumerate *enumerate = NULL;
    struct udev_list_entry *devices = NULL, *dev_list_entry = NULL;
    char *bootdisk = NULL;

    udev = udev_new();
    if (!udev) {
        fprintf(stderr, "Error: cannot create udev context\n");
        return 1;
    }

    bootdisk = get_boot_disk();
    if (bootdisk)
        printf("System boot disk (excluded): %s\n\n", bootdisk);
    else
        printf("Warning: boot disk not found automatically; proceed carefully.\n\n");

    /* Prepare JSON array to hold devices */
    struct json_object *jarr = json_object_new_array();

    enumerate = udev_enumerate_new(udev);
    udev_enumerate_add_match_subsystem(enumerate, "block");
    udev_enumerate_scan_devices(enumerate);

    devices = udev_enumerate_get_list_entry(enumerate);
    udev_list_entry_foreach(dev_list_entry, devices) {
        const char *path = udev_list_entry_get_name(dev_list_entry);
        struct udev_device *dev = udev_device_new_from_syspath(udev, path);
        if (!dev) continue;

        const char *devnode = udev_device_get_devnode(dev);
        const char *devtype = udev_device_get_devtype(dev);

        /* we want only top-level disks (not partitions) */
        if (!devnode || !devtype || strcmp(devtype, "disk") != 0) {
            udev_device_unref(dev);
            continue;
        }

        /* exclude boot disk */
        if (bootdisk && strcmp(devnode, bootdisk) == 0) {
            udev_device_unref(dev);
            continue;
        }

        /* gather attributes with fallbacks (different devices expose different names) */
        const char *model = get_attr_fallback(dev, "device/model", "model");
        const char *serial = get_attr_fallback(dev, "device/serial", "serial");
        const char *wwn = udev_device_get_property_value(dev, "ID_WWN");
        const char *bus  = udev_device_get_property_value(dev, "ID_BUS");
        const char *tran = udev_device_get_sysattr_value(dev, "device/transport");
        const char *size_blocks = udev_device_get_sysattr_value(dev, "size"); /* number of 512-byte sectors */

        unsigned long long blocks = 0, bytes = 0;
        if (size_blocks) {
            blocks = strtoull(size_blocks, NULL, 10);
            bytes = blocks * 512ULL;
        }

        /* Build JSON object for this device */
        struct json_object *jdev = json_object_new_object();
        json_object_object_add(jdev, "path", json_object_new_string(devnode));
        json_object_object_add(jdev, "model", json_object_new_string(model ? model : "Unknown"));
        json_object_object_add(jdev, "serial", json_object_new_string(serial ? serial : (wwn ? wwn : "Unknown")));
        json_object_object_add(jdev, "bus", json_object_new_string(bus ? bus : (tran ? tran : "Unknown")));
        json_object_object_add(jdev, "size_blocks", json_object_new_string(size_blocks ? size_blocks : "Unknown"));
        json_object_object_add(jdev, "size_bytes", json_object_new_int64(bytes));

        json_object_array_add(jarr, jdev);

        printf("Found: %-10s  Model: %-20s  Serial: %-20s  Bus: %-6s  Size: %llu bytes\n",
               devnode,
               model ? model : "Unknown",
               serial ? serial : (wwn ? wwn : "Unknown"),
               bus ? bus : (tran ? tran : "Unknown"),
               bytes);

        udev_device_unref(dev);
    }

    /* Write JSON file */
    const char *outfn = "device_inventory.json";
    FILE *out = fopen(outfn, "w");
    if (out) {
        const char *jstr = json_object_to_json_string_ext(jarr, JSON_C_TO_STRING_PRETTY);
        fputs(jstr, out);
        fclose(out);
        printf("\nSaved device inventory to %s\n", outfn);
    } else {
        fprintf(stderr, "Error: cannot write %s\n", outfn);
    }

    /* cleanup */
    json_object_put(jarr);
    udev_enumerate_unref(enumerate);
    udev_unref(udev);
    if (bootdisk) free(bootdisk);

    return 0;
}
