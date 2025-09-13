#include "wipe_ops.h"
#include "exec_cmd.h"
#include "util.h"

static void set_res(op_result_t *r, int code, const char *name, const char *out, int ec) {
    r->method_code = code;
    strncpy(r->method_name, name, sizeof(r->method_name)-1);
    r->exit_code = ec;
    if (out) strncpy(r->transcript, out, sizeof(r->transcript)-1);
}

int do_clear(const device_t *dev, bool dry_run, op_result_t *res) {
    char cmd[MAX_CMD], out[MAX_OUT]; int ec=0;

    int disc = is_blkdiscard_supported(dev->path);
    if (disc == 1) {
        snprintf(cmd, sizeof(cmd), "blkdiscard %s", dev->path);
        run_cmd_capture(cmd, out, sizeof(out), &ec, dry_run);
        set_res(res, 1, "blkdiscard", out, ec);
        snprintf(res->verify_note, sizeof(res->verify_note), "Issued full-device discard.");
        return 0;
    }

    // Fallback: single-pass zero overwrite
    snprintf(cmd, sizeof(cmd), "dd if=/dev/zero of=%s bs=16M status=progress conv=fdatasync", dev->path);
    run_cmd_capture(cmd, out, sizeof(out), &ec, dry_run);
    set_res(res, 2, "overwrite-1pass", out, ec);
    snprintf(res->verify_note, sizeof(res->verify_note), "Single-pass overwrite requested.");
    return 0;
}

int do_purge(const device_t *dev, bool dry_run, bool prefer_crypto, op_result_t *res) {
    char cmd[MAX_CMD], out[MAX_OUT]; int ec=0;

    if (dev->bus == BUS_NVME) {
        if (prefer_crypto) {
            snprintf(cmd, sizeof(cmd), "nvme sanitize %s --sanact=2", dev->path);
            run_cmd_capture(cmd, out, sizeof(out), &ec, dry_run);
            set_res(res, 10, "nvme-sanitize-crypto", out, ec);
        } else {
            snprintf(cmd, sizeof(cmd), "nvme sanitize %s --sanact=1", dev->path);
            run_cmd_capture(cmd, out, sizeof(out), &ec, dry_run);
            set_res(res, 11, "nvme-sanitize-block", out, ec);
        }
        return 0;
    }

    if (dev->bus == BUS_SATA || dev->bus == BUS_SAS) {
        // ATA Secure Erase: set password (NULL pass is common vendor quirk; use "p" minimal)
        // Real-world flows may require security freeze lock handling & power cycle.
        snprintf(cmd, sizeof(cmd), "hdparm --user-master u --security-set-pass p %s", dev->path);
        run_cmd_capture(cmd, out, sizeof(out), &ec, dry_run);

        snprintf(cmd, sizeof(cmd), "hdparm --security-erase p %s", dev->path);
        run_cmd_capture(cmd, out, sizeof(out), &ec, dry_run);
        set_res(res, 20, "ata-secure-erase", out, ec);
        return 0;
    }

    // Unknown bus: fallback to clear
    return do_clear(dev, dry_run, res);
}
