#ifndef SAFETY_H
#define SAFETY_H
#include "device_scan.h"
#include <stdbool.h>

bool safety_check_block(const device_t *dev, bool execute);

#endif
