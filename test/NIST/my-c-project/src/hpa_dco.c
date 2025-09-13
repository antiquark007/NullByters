#include "hpa_dco.h"
#include "exec_cmd.h"
#include "util.h"
#include <string.h>

int check_hpa_dco(const device_t *dev, hpa_dco_report_t *rep) {
    memset(rep, 0, sizeof(*rep));
    if (dev->bus != BUS_SATA && dev->bus != BUS_SAS) return 0; // NVMe doesn't use HPA/DCO

    // hdparm -N /dev/sdX gives max/native
    char cmd[MAX_CMD], out[MAX_OUT]; int ec=0;
    snprintf(cmd, sizeof(cmd), "hdparm -N %s 2>/dev/null", dev->path);
    if (run_cmd_capture(cmd, out, sizeof(out), &ec, false)==0 && ec==0) {
        // Parse e.g.: "max sectors   = 586070255/586072368, HPA is enabled"
        if (strstr(out, "HPA")) rep->hpa_present = (strstr(out, "enabled") != NULL);
        unsigned long long a=0,b=0;
        if (sscanf(out, "%*[^=]= %llu/%llu", &a, &b)==2) {
            rep->hpa_max = a; rep->native_max = b;
        }
        if (rep->hpa_present && rep->native_max > 0) {
            snprintf(rep->suggest_cmd, sizeof(rep->suggest_cmd),
                     "hdparm -N p%llu %s", rep->native_max, dev->path);
        }
    }

    // DCO identify
    memset(out, 0, sizeof(out));
    snprintf(cmd, sizeof(cmd), "hdparm --dco-identify %s 2>/dev/null", dev->path);
    if (run_cmd_capture(cmd, out, sizeof(out), &ec, false)==0 && ec==0) {
        // If output contains "DCO feature set", assume DCO present; detecting limitation is vendor-specific.
        if (strstr(out, "DCO")) {
            // Heuristic: if HPA suggests reduced size OR DCO report hints limitation
            if (rep->hpa_present || strstr(out, "word")) {
                rep->dco_limited = true; // conservative
            }
        }
    }

    return 0;
}

int force_restore_hpa_dco(const device_t *dev, bool dry_run) {
    if (dev->bus != BUS_SATA && dev->bus != BUS_SAS) return 0;
    char cmd[MAX_CMD], out[MAX_OUT]; int ec=0;

    // Restore DCO first (if needed)
    snprintf(cmd, sizeof(cmd), "hdparm --yes-i-know-what-i-am-doing --dco-restore %s", dev->path);
    run_cmd_capture(cmd, out, sizeof(out), &ec, dry_run);

    // Set max to native (HPA removal)
    snprintf(cmd, sizeof(cmd), "hdparm -N %s | awk -F'[/, ]+' '/max sectors/ {print $5}'", dev->path);
    if (run_cmd_capture(cmd, out, sizeof(out), &ec, false)==0 && ec==0) {
        unsigned long long native = strtoull(out, NULL, 10);
        if (native > 0) {
            snprintf(cmd, sizeof(cmd), "hdparm -N p%llu %s", native, dev->path);
            run_cmd_capture(cmd, out, sizeof(out), &ec, dry_run);
        }
    }
    return 0;
}
