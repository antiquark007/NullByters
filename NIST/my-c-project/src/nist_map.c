#include "nist_map.h"

const char* nist_level_for(const char *bus_name, wipe_mode_t mode, int method_code) {
    // Simple mapping aligned to SP 800-88 categories
    switch (method_code) {
        case 10: // NVMe sanitize crypto
        case 11: // NVMe sanitize block
        case 20: // ATA secure erase
            return "purge";
        case 1:  // blkdiscard
            // Most SSDs: discard of all LBAs qualifies as purge if controller guarantees purge semantics.
            // Conservatively: mark purge for NVMe discard-of-all; else clear.
            if (bus_name && strstr(bus_name, "NVMe")) return "purge";
            return "clear";
        case 2:  // overwrite single pass
        default:
            return "clear";
    }
}
