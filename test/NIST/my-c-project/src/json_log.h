#ifndef JSON_LOG_H
#define JSON_LOG_H
#include "device_scan.h"
#include "hpa_dco.h"
#include "wipe_ops.h"
#include "nist_map.h"
#include <stdio.h>

int write_json_cert(FILE *fp,
                    const char *tool_name,
                    const char *version,
                    const device_t *dev,
                    const hpa_dco_report_t *hpa,
                    wipe_mode_t mode,
                    const op_result_t *res,
                    const char *started_iso8601,
                    const char *finished_iso8601);

#endif
