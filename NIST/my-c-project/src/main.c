#include "device_scan.h"
#include "safety.h"
#include "hpa_dco.h"
#include "wipe_ops.h"
#include "json_log.h"
#include "util.h"
#include <time.h>

static void iso_time(char *buf, size_t n) {
    time_t t = time(NULL);
    struct tm tmv;
    gmtime_r(&t, &tmv);
    strftime(buf, n, "%Y-%m-%dT%H:%M:%SZ", &tmv);
}

static void usage() {
    printf("cwipe - safe data sanitization core (Linux)\n");
    printf("Usage:\n");
    printf("  cwipe --scan\n");
    printf("  cwipe --device DEV --mode [clear|purge] [--execute] [--prefer-crypto] [--check-hpa] [--force-hpa] [--out FILE]\n");
    printf("\nDefaults: dry-run unless --execute is provided.\n");
}

int main(int argc, char **argv) {
    const char *devpath = NULL;
    const char *mode_s = NULL;
    bool execute = false;
    bool prefer_crypto = true;
    bool check_hpa_only = false;
    bool force_hpa = false;
    const char *outf = NULL;

    for (int i=1; i<argc; ++i) {
        if (!strcmp(argv[i], "--device") && i+1<argc) devpath = argv[++i];
        else if (!strcmp(argv[i], "--mode") && i+1<argc) mode_s = argv[++i];
        else if (!strcmp(argv[i], "--execute")) execute = true;
        else if (!strcmp(argv[i], "--prefer-crypto")) prefer_crypto = true;
        else if (!strcmp(argv[i], "--no-crypto")) prefer_crypto = false;
        else if (!strcmp(argv[i], "--check-hpa")) check_hpa_only = true;
        else if (!strcmp(argv[i], "--force-hpa")) force_hpa = true;
        else if (!strcmp(argv[i], "--out") && i+1<argc) outf = argv[++i];
        else if (!strcmp(argv[i], "--scan")) { mode_s = "scan"; }
        else if (!strcmp(argv[i], "--help") || !strcmp(argv[i], "-h")) { usage(); return 0; }
    }

    if (!mode_s) { usage(); return 1; }

    if (!strcmp(mode_s, "scan")) {
        // Minimal scan dump: rely on lsblk JSON for convenience
        char out[MAX_OUT]; int ec=0;
        run_cmd_capture("lsblk -J -o NAME,TYPE,SIZE,MODEL,SERIAL,TRAN", out, sizeof(out), &ec, false);
        printf("%s\n", out);
        return 0;
    }

    if (!devpath) { die("Missing --device"); }

    device_t dev;
    detect_device(devpath, &dev);

    if (!safety_check_block(&dev, execute)) return 2;

    hpa_dco_report_t rep;
    check_hpa_dco(&dev, &rep);

    if (check_hpa_only) {
        // Print a small JSON with findings
        printf("{\"device\":\"%s\",\"hpa_present\":%s,\"hpa_max\":%llu,\"native_max\":%llu,"
               "\"dco_limited\":%s,\"suggest\":\"%s\"}\n",
               dev.path, rep.hpa_present?"true":"false",
               rep.hpa_max, rep.native_max,
               rep.dco_limited?"true":"false",
               rep.suggest_cmd);
        return 0;
    }

    if (force_hpa && execute) {
        warn("FORCE-HPA/DCO requested. Proceeding with extreme caution.");
        force_restore_hpa_dco(&dev, false); // we still want real status output lines even when executing
    } else if (force_hpa && !execute) {
        // dry-run path for force-hpa preview
        force_restore_hpa_dco(&dev, true);
    }

    wipe_mode_t mode = MODE_CLEAR;
    if (!strcmp(mode_s, "purge")) mode = MODE_PURGE;
    else if (!strcmp(mode_s, "clear")) mode = MODE_CLEAR;
    else die("Unknown --mode (use clear|purge|scan)");

    op_result_t res = {0};
    char t_start[32], t_end[32];
    iso_time(t_start, sizeof(t_start));

    if (mode == MODE_CLEAR) {
        do_clear(&dev, !execute, &res);
    } else {
        do_purge(&dev, !execute, prefer_crypto, &res);
    }

    iso_time(t_end, sizeof(t_end));

    FILE *fp = stdout;
    if (outf) {
        fp = fopen(outf, "w");
        if (!fp) die("Failed to open output file");
    }

    write_json_cert(fp, "cwipe", "0.1.0", &dev, &rep, mode, &res, t_start, t_end);

    if (fp != stdout) fclose(fp);

    if (!execute) {
        info("NOTE: This was a dry-run. Use --execute to perform the actual operation.");
    }

    return res.exit_code;
}
