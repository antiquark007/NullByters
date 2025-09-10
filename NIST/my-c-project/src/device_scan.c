#define _XOPEN_SOURCE 700
#include "device_scan.h"
#include "exec_cmd.h"
#include "util.h"
#include <sys/stat.h>

static int file_exists(const char *p) {
    struct stat st;
    return stat(p, &st) == 0;
}

static void trim(char *s) {
    size_t n = strlen(s);
    while (n && (s[n-1]=='\n' || s[n-1]=='\r' || s[n-1]==' ' || s[n-1]=='\t')) { s[--n]=0; }
}

int is_nvme(const char *devpath) {
    return strstr(devpath, "nvme") != NULL;
}

int is_sata_like(const char *devpath) {
    return strstr(devpath, "sd") != NULL; // crude but effective for /dev/sdX
}

static bus_t guess_bus(const char *devpath) {
    if (is_nvme(devpath)) return BUS_NVME;
    if (is_sata_like(devpath)) return BUS_SATA; // could be USB-SATA bridge
    return BUS_UNKNOWN;
}

static int read_sysfs_string(const char *sys_path, char *buf, int cap) {
    FILE *f = fopen(sys_path, "r");
    if (!f) return -1;
    int n = fread(buf, 1, cap-1, f);
    buf[n] = 0;
    fclose(f);
    trim(buf);
    return 0;
}

static uint64_t read_size_bytes_lsblk(const char *devpath) {
    char cmd[MAX_CMD], out[MAX_OUT]; int ec=0;
    snprintf(cmd, sizeof(cmd), "lsblk -nb -o SIZE %s 2>/dev/null | head -n1", devpath);
    if (run_cmd_capture(cmd, out, sizeof(out), &ec, false) == 0 && ec==0) {
        return strtoull(out, NULL, 10);
    }
    return 0;
}

static void detect_system_guard(device_t *d) {
    // crude: check if / is on same disk
    char out[MAX_OUT]; int ec=0;
    // Find root mount device
    if (run_cmd_capture("findmnt -n -o SOURCE /", out, sizeof(out), &ec, false)==0 && ec==0) {
        // Might be /dev/sdaX or UUID=...
        if (strstr(out, "/dev/")) {
            char *p = strstr(out, "/dev/");
            char rootdev[64]={0};
            sscanf(p, "%63s", rootdev);
            // Normalize: for /dev/sda1 -> /dev/sda
            char base[64]={0};
            if (strstr(rootdev, "nvme")) {
                // /dev/nvme0n1p1 -> /dev/nvme0n1
                char *pp = strstr(rootdev, "p");
                if (pp) *pp = 0;
                strncpy(base, rootdev, sizeof(base)-1);
            } else {
                // strip partition digit(s)
                strncpy(base, rootdev, sizeof(base)-1);
                for (int i=strlen(base)-1; i>=0; --i) {
                    if (base[i] >= '0' && base[i] <= '9') base[i]=0; else break;
                }
            }
            if (strcmp(base, d->path)==0) d->is_system_device = true;
        }
    }
}

int detect_device(const char *devpath, device_t *out) {
    memset(out, 0, sizeof(*out));
    strncpy(out->path, devpath, sizeof(out->path)-1);
    out->bus = guess_bus(devpath);

    // Model/serial/firmware via udevadm as fallback
    char cmd[MAX_CMD], outbuf[MAX_OUT]; int ec=0;

    snprintf(cmd, sizeof(cmd), "udevadm info --query=property --name=%s 2>/dev/null", devpath);
    if (run_cmd_capture(cmd, outbuf, sizeof(outbuf), &ec, false)==0 && ec==0) {
        // Parse
        char *line = strtok(outbuf, "\n");
        while (line) {
            if (strncmp(line, "ID_MODEL=", 9)==0) strncpy(out->model, line+9, sizeof(out->model)-1);
            if (strncmp(line, "ID_SERIAL_SHORT=", 16)==0) strncpy(out->serial, line+16, sizeof(out->serial)-1);
            if (strncmp(line, "ID_REVISION=", 12)==0) strncpy(out->firmware, line+12, sizeof(out->firmware)-1);
            if (strncmp(line, "ID_BUS=", 7)==0) {
                if (strstr(line, "nvme")) out->bus = BUS_NVME;
                else if (strstr(line, "ata")||strstr(line,"scsi")) out->bus = BUS_SATA;
                else if (strstr(line, "usb")) out->bus = BUS_USB;
            }
            line = strtok(NULL, "\n");
        }
    }

    // Size
    out->size_bytes = read_size_bytes_lsblk(devpath);

    detect_system_guard(out);

    return 0;
}

int is_blkdiscard_supported(const char *devpath) {
    // We assume if fstrim/discard is supported: query via blkdiscard --help? Not reliable.
    // Safer: attempt a non-destructive check via "lsblk -D" (discard granularity > 0)
    char cmd[MAX_CMD], out[MAX_OUT]; int ec=0;
    snprintf(cmd, sizeof(cmd), "lsblk -Dno DISC-GRAN %s 2>/dev/null", devpath);
    if (run_cmd_capture(cmd, out, sizeof(out), &ec, false)==0 && ec==0) {
        long long gran = atoll(out);
        return (gran > 0) ? 1 : 0;
    }
    return -1;
}
