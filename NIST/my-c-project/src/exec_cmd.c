#include "exec_cmd.h"
#include "util.h"
#include <stdio.h>

int run_cmd_capture(const char *cmd, char *outbuf, int outcap, int *exit_code, bool dry_run) {
    if (dry_run) {
        snprintf(outbuf, outcap, "DRY-RUN: %s\n", cmd);
        if (exit_code) *exit_code = 0;
        fprintf(stderr, "DRY-RUN would run: %s\n", cmd);
        return 0;
    }
    FILE *fp = popen(cmd, "r");
    if (!fp) {
        if (exit_code) *exit_code = 127;
        return -1;
    }
    int n = fread(outbuf, 1, outcap-1, fp);
    outbuf[n] = '\0';
    int rc = pclose(fp);
    if (exit_code) *exit_code = WEXITSTATUS(rc);
    return 0;
}
