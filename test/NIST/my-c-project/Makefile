CC?=gcc
CFLAGS=-O2 -Wall -Wextra -std=c11

SRC=src/main.c src/exec_cmd.c src/device_scan.c src/safety.c src/wipe_ops.c src/hpa_dco.c src/json_log.c src/nist_map.c
HDR=src/util.h src/exec_cmd.h src/device_scan.h src/safety.h src/wipe_ops.h src/hpa_dco.h src/json_log.h src/nist_map.h

cwipe: $(SRC) $(HDR)
    $(CC) $(CFLAGS) -o $@ $(SRC)

clean:
    rm -f cwipe
