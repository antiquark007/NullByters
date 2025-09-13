#ifndef UTIL_H
#define UTIL_H

#define _GNU_SOURCE
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <stdbool.h>

#define MAX_PATH 256
#define MAX_CMD  1024
#define MAX_OUT  8192

static inline void die(const char *msg) {
    fprintf(stderr, "ERROR: %s\n", msg);
    exit(1);
}

static inline void warn(const char *msg) {
    fprintf(stderr, "WARN: %s\n", msg);
}

static inline void info(const char *msg) {
    fprintf(stderr, "INFO: %s\n", msg);
}

#endif
