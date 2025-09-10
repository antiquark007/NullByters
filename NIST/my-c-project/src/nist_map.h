#ifndef NIST_MAP_H
#define NIST_MAP_H

typedef enum { MODE_CLEAR, MODE_PURGE } wipe_mode_t;

const char* nist_level_for(const char *bus_name, wipe_mode_t mode, int method_code);
/*
 method_code examples:
  1 = blkdiscard
  2 = overwrite-1pass
 10 = nvme-sanitize-crypto
 11 = nvme-sanitize-block
 20 = ata-secure-erase
*/

#endif
