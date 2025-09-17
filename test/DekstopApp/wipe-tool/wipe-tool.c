#include <stdio.h>
#include <stdlib.h>
#include <string.h>

void wipe_device(const char *device, const char *method) {
    char command[256];

    if (strcmp(method, "Zero Fill") == 0) {
        snprintf(command, sizeof(command), "dd if=/dev/zero of=%s bs=1M status=progress", device);
    } else if (strcmp(method, "Random Fill") == 0) {
        snprintf(command, sizeof(command), "dd if=/dev/urandom of=%s bs=1M status=progress", device);
    } else if (strcmp(method, "Shred + Zero") == 0) {
        snprintf(command, sizeof(command), "shred -v -n 3 %s && dd if=/dev/zero of=%s bs=1M status=progress", device, device);
    } else {
        fprintf(stderr, "Unknown wipe method selected.\n");
        return;
    }

    printf("Wiping device %s using method: %s\n", device, method);
    int result = system(command);
    if (result == -1) {
        fprintf(stderr, "Failed to execute wipe command.\n");
    } else {
        printf("Wipe operation completed successfully.\n");
    }
}

int main(int argc, char *argv[]) {
    if (argc != 3) {
        fprintf(stderr, "Usage: %s <device> <method>\n", argv[0]);
        return EXIT_FAILURE;
    }

    const char *device = argv[1];
    const char *method = argv[2];

    wipe_device(device, method);
    return EXIT_SUCCESS;
}