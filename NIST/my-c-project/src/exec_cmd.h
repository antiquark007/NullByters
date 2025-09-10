#ifndef EXEC_CMD_H
#define EXEC_CMD_H
#include <stdbool.h>

int run_cmd_capture(const char *cmd, char *outbuf, int outcap, int *exit_code, bool dry_run);

#endif
