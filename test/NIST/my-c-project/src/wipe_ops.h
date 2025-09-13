#ifndef WIPE_OPS_H
#define WIPE_OPS_H
#include "device_scan.h"
#include "nist_map.h"
#include <stdbool.h>

typedef struct {
    int method_code;       // see nist_map.h
    char method_name[64];  // "nvme-sanitize-crypto", "ata-secure-erase", "blkdiscard", "overwrite-1pass"
    int exit_code;
    char transcript[2048]; // captured stdout/stderr (trimmed)
    char verify_note[256];
} op_result_t;

int do_clear(const device_t *dev, bool dry_run, op_result_t *res);
int do_purge(const device_t *dev, bool dry_run, bool prefer_crypto, op_result_t *res);

#endif
