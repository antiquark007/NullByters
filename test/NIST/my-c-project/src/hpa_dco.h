#ifndef HPA_DCO_H
#define HPA_DCO_H
#include "device_scan.h"
#include <stdbool.h>

typedef struct {
    bool hpa_present;
    unsigned long long hpa_max;
    unsigned long long native_max;
    bool dco_limited;
    char suggest_cmd[512];
} hpa_dco_report_t;

int check_hpa_dco(const device_t *dev, hpa_dco_report_t *rep);
int force_restore_hpa_dco(const device_t *dev, bool dry_run);

#endif
