#ifndef DEVICE_SCAN_H
#define DEVICE_SCAN_H
#include <stdbool.h>
#include <stdint.h>

typedef enum { BUS_UNKNOWN, BUS_SATA, BUS_NVME, BUS_USB, BUS_SAS } bus_t;

typedef struct {
    char path[MAX_PATH];     // /dev/sdX or /dev/nvme0n1
    char model[256];
    char serial[256];
    char firmware[128];
    uint64_t size_bytes;
    bus_t bus;
    bool is_system_device;   // boot/system disk guard
} device_t;

int detect_device(const char *devpath, device_t *out);
int is_blkdiscard_supported(const char *devpath); // 0=no, 1=yes, <0=unknown
int is_nvme(const char *devpath);
int is_sata_like(const char *devpath);

#endif
