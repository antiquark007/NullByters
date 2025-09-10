#include "safety.h"
#include "util.h"
#include <stdio.h>

bool safety_check_block(const device_t *dev, bool execute) {
    if (dev->is_system_device && execute) {
        fprintf(stderr, "REFUSE: %s appears to be the system/boot disk. Aborting.\n", dev->path);
        return false;
    }
    return true;
}
