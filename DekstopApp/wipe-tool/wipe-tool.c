#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <fcntl.h>
#include <errno.h>
#include <time.h>
#include <sys/stat.h>
#include <sys/ioctl.h>
#include <getopt.h>

#ifdef _WIN32
#include <windows.h>
#include <winioctl.h>
#else
#include <linux/fs.h>
#include <sys/mount.h>
#endif

#define BUFFER_SIZE (1024 * 1024) // 1MB buffer
#define MAX_PATH_LEN 512
#define MAX_DEVICES 32

typedef struct {
    char path[MAX_PATH_LEN];
    char name[256];
    char serial[128];
    long long size_bytes;
    double size_gb;
} device_info_t;

typedef struct {
    device_info_t devices[MAX_DEVICES];
    int count;
} device_list_t;

typedef enum {
    WIPE_CLEAR,
    WIPE_PURGE,
    WIPE_DESTROY
} wipe_method_t;

// Function prototypes
int scan_devices(device_list_t *list);
int wipe_device(const char *device_path, wipe_method_t method, const char *output_log);
void print_progress(int percent, const char *message);
void create_wipe_log(const char *device_path, wipe_method_t method, const char *output_file, int success);
int is_system_drive(const char *path);
long long get_device_size(const char *device_path);
void print_devices_json(device_list_t *devices);

// Cross-platform device scanning
int scan_devices(device_list_t *list) {
    list->count = 0;
    
#ifdef _WIN32
    // Windows device enumeration
    char drive_strings[256];
    DWORD drives = GetLogicalDriveStrings(sizeof(drive_strings), drive_strings);
    
    char *drive = drive_strings;
    while (*drive) {
        if (GetDriveType(drive) == DRIVE_REMOVABLE) {
            device_info_t *dev = &list->devices[list->count];
            strncpy(dev->path, drive, sizeof(dev->path) - 1);
            
            ULARGE_INTEGER free_bytes, total_bytes;
            if (GetDiskFreeSpaceEx(drive, &free_bytes, &total_bytes, NULL)) {
                dev->size_bytes = total_bytes.QuadPart;
                dev->size_gb = (double)total_bytes.QuadPart / (1024.0 * 1024.0 * 1024.0);
            }
            
            // Get volume information
            char volume_name[256];
            if (GetVolumeInformation(drive, volume_name, sizeof(volume_name), 
                                   NULL, NULL, NULL, NULL, 0)) {
                snprintf(dev->name, sizeof(dev->name), "%s (%s)", volume_name, drive);
            } else {
                snprintf(dev->name, sizeof(dev->name), "Removable Drive (%s)", drive);
            }
            
            snprintf(dev->serial, sizeof(dev->serial), "WIN_%d", list->count);
            list->count++;
        }
        drive += strlen(drive) + 1;
    }
#else
    // Linux device enumeration
    FILE *fp = fopen("/proc/partitions", "r");
    if (!fp) return -1;
    
    char line[256];
    fgets(line, sizeof(line), fp); // skip header
    fgets(line, sizeof(line), fp); // skip header
    
    while (fgets(line, sizeof(line), fp) && list->count < MAX_DEVICES) {
        int major, minor;
        long long blocks;
        char name[64];
        
        if (sscanf(line, "%d %d %lld %s", &major, &minor, &blocks, name) == 4) {
            // Filter for removable devices (USB, etc.)
            char sys_path[256];
            snprintf(sys_path, sizeof(sys_path), "/sys/block/%s/removable", name);
            
            FILE *removable_fp = fopen(sys_path, "r");
            if (removable_fp) {
                char removable[8];
                if (fgets(removable, sizeof(removable), removable_fp) && removable[0] == '1') {
                    device_info_t *dev = &list->devices[list->count];
                    snprintf(dev->path, sizeof(dev->path), "/dev/%s", name);
                    dev->size_bytes = blocks * 1024;
                    dev->size_gb = (double)(blocks * 1024) / (1024.0 * 1024.0 * 1024.0);
                    
                    // Try to get device model
                    snprintf(sys_path, sizeof(sys_path), "/sys/block/%s/device/model", name);
                    FILE *model_fp = fopen(sys_path, "r");
                    if (model_fp) {
                        if (fgets(dev->name, sizeof(dev->name), model_fp)) {
                            // Remove newline
                            char *nl = strchr(dev->name, '\n');
                            if (nl) *nl = '\0';
                        }
                        fclose(model_fp);
                    } else {
                        snprintf(dev->name, sizeof(dev->name), "USB Device %s", name);
                    }
                    
                    // Try to get serial
                    snprintf(sys_path, sizeof(sys_path), "/sys/block/%s/device/serial", name);
                    FILE *serial_fp = fopen(sys_path, "r");
                    if (serial_fp) {
                        if (fgets(dev->serial, sizeof(dev->serial), serial_fp)) {
                            char *nl = strchr(dev->serial, '\n');
                            if (nl) *nl = '\0';
                        }
                        fclose(serial_fp);
                    } else {
                        snprintf(dev->serial, sizeof(dev->serial), "UNKNOWN_%d", list->count);
                    }
                    
                    list->count++;
                }
                fclose(removable_fp);
            }
        }
    }
    fclose(fp);
#endif
    
    return list->count;
}

// Check if device is a system drive
int is_system_drive(const char *path) {
    if (!path) return 1;
    
#ifdef _WIN32
    // Block C: drive and system paths
    if (strncmp(path, "C:", 2) == 0 || strncmp(path, "c:", 2) == 0) return 1;
    if (strstr(path, "Windows") || strstr(path, "WINDOWS")) return 1;
#else
    // Block root filesystem and common system mounts
    if (strcmp(path, "/") == 0) return 1;
    if (strncmp(path, "/dev/sd", 7) == 0) {
        // Additional checks for system drives
        char check_path[256];
        snprintf(check_path, sizeof(check_path), "%s1", path);
        
        FILE *mounts = fopen("/proc/mounts", "r");
        if (mounts) {
            char line[512];
            while (fgets(line, sizeof(line), mounts)) {
                if (strstr(line, check_path) && (strstr(line, " / ") || strstr(line, " /boot"))) {
                    fclose(mounts);
                    return 1;
                }
            }
            fclose(mounts);
        }
    }
#endif
    return 0;
}

// Get device size
long long get_device_size(const char *device_path) {
#ifdef _WIN32
    HANDLE handle = CreateFile(device_path, GENERIC_READ, 
                              FILE_SHARE_READ | FILE_SHARE_WRITE, 
                              NULL, OPEN_EXISTING, 0, NULL);
    if (handle == INVALID_HANDLE_VALUE) return -1;
    
    LARGE_INTEGER size;
    if (!GetFileSizeEx(handle, &size)) {
        CloseHandle(handle);
        return -1;
    }
    CloseHandle(handle);
    return size.QuadPart;
#else
    int fd = open(device_path, O_RDONLY);
    if (fd < 0) return -1;
    
    long long size = 0;
    if (ioctl(fd, BLKGETSIZE64, &size) < 0) {
        close(fd);
        return -1;
    }
    close(fd);
    return size;
#endif
}

// Print progress in JSON format for GUI
void print_progress(int percent, const char *message) {
    printf("{\"progress\": %d, \"message\": \"%s\"}\n", percent, message ? message : "");
    fflush(stdout);
}

// Simple JSON output without cjson library
void print_devices_json(device_list_t *devices) {
    printf("{\n  \"devices\": [\n");
    for (int i = 0; i < devices->count; i++) {
        printf("    {\n");
        printf("      \"name\": \"%s\",\n", devices->devices[i].name);
        printf("      \"path\": \"%s\",\n", devices->devices[i].path);
        printf("      \"serial\": \"%s\",\n", devices->devices[i].serial);
        printf("      \"size_gb\": %.1f\n", devices->devices[i].size_gb);
        printf("    }%s\n", (i < devices->count - 1) ? "," : "");
    }
    printf("  ]\n}\n");
}

// Secure wipe patterns
static const unsigned char PATTERN_ZEROS[256] = {0};
static const unsigned char PATTERN_ONES[256] = {[0 ... 255] = 0xFF};
static const unsigned char PATTERN_RANDOM[256] = {
    0x55, 0xAA, 0x33, 0xCC, 0x0F, 0xF0, 0x99, 0x66,
    0x12, 0x34, 0x56, 0x78, 0x9A, 0xBC, 0xDE, 0xF0,
    0x11, 0x22, 0x44, 0x88, 0x10, 0x20, 0x40, 0x80,
    0xA5, 0x5A, 0xC3, 0x3C, 0x69, 0x96, 0x87, 0x78,
    0x15, 0x2A, 0x54, 0xA8, 0x51, 0xA2, 0x45, 0x8A,
    0x35, 0x6A, 0xD4, 0xA9, 0x53, 0xA6, 0x4D, 0x9A,
    0x25, 0x4A, 0x94, 0x29, 0x52, 0xA4, 0x49, 0x92,
    0x65, 0xCA, 0x95, 0x2B, 0x56, 0xAC, 0x59, 0xB2,
    // Continue pattern for full 256 bytes
    [64 ... 255] = 0x5A
};

// Main wipe function
int wipe_device(const char *device_path, wipe_method_t method, const char *output_log) {
    if (is_system_drive(device_path)) {
        fprintf(stderr, "ERROR: Refusing to wipe system drive: %s\n", device_path);
        return -1;
    }
    
    long long device_size = get_device_size(device_path);
    if (device_size <= 0) {
        fprintf(stderr, "ERROR: Cannot determine device size: %s\n", device_path);
        return -1;
    }
    
    int fd;
#ifdef _WIN32
    fd = _open(device_path, _O_WRONLY | _O_BINARY);
#else
    fd = open(device_path, O_WRONLY | O_SYNC);
#endif
    
    if (fd < 0) {
        fprintf(stderr, "ERROR: Cannot open device %s: %s\n", device_path, strerror(errno));
        return -1;
    }
    
    unsigned char *buffer = malloc(BUFFER_SIZE);
    if (!buffer) {
        close(fd);
        return -1;
    }
    
    long long total_written = 0;
    int passes = 1;
    int success = 1;
    
    // Determine number of passes based on method
    switch (method) {
        case WIPE_CLEAR: passes = 1; break;
        case WIPE_PURGE: passes = 3; break;
        case WIPE_DESTROY: passes = 7; break;
    }
    
    print_progress(0, "Starting secure wipe...");
    
    for (int pass = 0; pass < passes && success; pass++) {
        // Select pattern for this pass
        const unsigned char *pattern;
        switch (pass % 3) {
            case 0: pattern = PATTERN_ZEROS; break;
            case 1: pattern = PATTERN_ONES; break;
            case 2: pattern = PATTERN_RANDOM; break;
            default: pattern = PATTERN_ZEROS; break;
        }
        
        // Fill buffer with pattern
        for (int i = 0; i < BUFFER_SIZE; i += 256) {
            memcpy(buffer + i, pattern, 256);
        }
        
        // Reset to beginning of device
#ifdef _WIN32
        _lseek(fd, 0, SEEK_SET);
#else
        lseek(fd, 0, SEEK_SET);
#endif
        
        total_written = 0;
        char progress_msg[256];
        snprintf(progress_msg, sizeof(progress_msg), "Pass %d/%d", pass + 1, passes);
        
        while (total_written < device_size) {
            long long remaining = device_size - total_written;
            size_t write_size = (remaining < BUFFER_SIZE) ? (size_t)remaining : BUFFER_SIZE;
            
#ifdef _WIN32
            int written = _write(fd, buffer, write_size);
#else
            ssize_t written = write(fd, buffer, write_size);
#endif
            
            if (written <= 0) {
                fprintf(stderr, "ERROR: Write failed at offset %lld: %s\n", 
                       total_written, strerror(errno));
                success = 0;
                break;
            }
            
            total_written += written;
            int percent = (int)((total_written * 100) / device_size);
            percent = (percent + (pass * 100)) / passes; // Adjust for multiple passes
            
            print_progress(percent, progress_msg);
        }
        
        // Force sync
#ifdef _WIN32
        _commit(fd);
#else
        fsync(fd);
#endif
    }
    
    if (success) {
        print_progress(100, "Wipe completed successfully");
    }
    
    close(fd);
    free(buffer);
    
    // Create log file
    create_wipe_log(device_path, method, output_log, success);
    
    return success ? 0 : -1;
}

// Simple JSON log creation without cjson library
void create_wipe_log(const char *device_path, wipe_method_t method, const char *output_file, int success) {
    FILE *fp = fopen(output_file, "w");
    if (!fp) return;
    
    time_t now = time(NULL);
    char timestamp[64];
    strftime(timestamp, sizeof(timestamp), "%Y-%m-%dT%H:%M:%SZ", gmtime(&now));
    
    const char *method_str = (method == WIPE_PURGE) ? "purge" : 
                            (method == WIPE_DESTROY) ? "destroy" : "clear";
    
    long long size = get_device_size(device_path);
    
    fprintf(fp, "{\n");
    fprintf(fp, "  \"device\": {\n");
    fprintf(fp, "    \"path\": \"%s\",\n", device_path);
    fprintf(fp, "    \"size_bytes\": %lld,\n", size);
    fprintf(fp, "    \"size_gb\": %.1f\n", (double)size / (1024.0 * 1024.0 * 1024.0));
    fprintf(fp, "  },\n");
    fprintf(fp, "  \"wipe\": {\n");
    fprintf(fp, "    \"method\": \"%s\",\n", method_str);
    fprintf(fp, "    \"nist_level\": \"%s\",\n", (method == WIPE_PURGE) ? "purge" : "clear");
    fprintf(fp, "    \"status\": \"%s\",\n", success ? "success" : "failed");
    fprintf(fp, "    \"started_at\": \"%s\",\n", timestamp);
    fprintf(fp, "    \"finished_at\": \"%s\"\n", timestamp);
    fprintf(fp, "  },\n");
    fprintf(fp, "  \"system\": {\n");
    fprintf(fp, "    \"tool_version\": \"1.0.0\",\n");
    fprintf(fp, "    \"platform\": \"%s\"\n", 
#ifdef _WIN32
            "Windows"
#else
            "Linux"
#endif
    );
    fprintf(fp, "  }\n");
    fprintf(fp, "}\n");
    
    fclose(fp);
}

// Main function
int main(int argc, char *argv[]) {
    int opt;
    int list_devices = 0;
    int json_output = 0;
    char device_path[MAX_PATH_LEN] = {0};
    char output_file[MAX_PATH_LEN] = "wipe_log.json";
    wipe_method_t method = WIPE_CLEAR;
    
    static struct option long_options[] = {
        {"list", no_argument, 0, 'l'},
        {"json", no_argument, 0, 'j'},
        {"device", required_argument, 0, 'd'},
        {"method", required_argument, 0, 'm'},
        {"output", required_argument, 0, 'o'},
        {"help", no_argument, 0, 'h'},
        {0, 0, 0, 0}
    };
    
    while ((opt = getopt_long(argc, argv, "ljd:m:o:h", long_options, NULL)) != -1) {
        switch (opt) {
            case 'l':
                list_devices = 1;
                break;
            case 'j':
                json_output = 1;
                break;
            case 'd':
                strncpy(device_path, optarg, sizeof(device_path) - 1);
                break;
            case 'm':
                if (strcmp(optarg, "clear") == 0) method = WIPE_CLEAR;
                else if (strcmp(optarg, "purge") == 0) method = WIPE_PURGE;
                else if (strcmp(optarg, "destroy") == 0) method = WIPE_DESTROY;
                else {
                    fprintf(stderr, "Invalid method: %s\n", optarg);
                    return 1;
                }
                break;
            case 'o':
                strncpy(output_file, optarg, sizeof(output_file) - 1);
                break;
            case 'h':
                printf("Usage: %s [OPTIONS]\n", argv[0]);
                printf("Options:\n");
                printf("  -l, --list          List available devices\n");
                printf("  -j, --json          Output in JSON format\n");
                printf("  -d, --device PATH   Device to wipe\n");
                printf("  -m, --method METHOD Wipe method (clear/purge/destroy)\n");
                printf("  -o, --output FILE   Output log file\n");
                printf("  -h, --help          Show this help\n");
                return 0;
            default:
                return 1;
        }
    }
    
    if (list_devices) {
        device_list_t devices;
        if (scan_devices(&devices) < 0) {
            fprintf(stderr, "ERROR: Failed to scan devices\n");
            return 1;
        }
        
        if (json_output) {
            print_devices_json(&devices);
        } else {
            printf("Available devices:\n");
            for (int i = 0; i < devices.count; i++) {
                printf("  %s: %s (%.1f GB, S/N: %s)\n", 
                       devices.devices[i].path,
                       devices.devices[i].name,
                       devices.devices[i].size_gb,
                       devices.devices[i].serial);
            }
        }
        return 0;
    }
    
    if (strlen(device_path) == 0) {
        fprintf(stderr, "ERROR: No device specified. Use --device option.\n");
        return 1;
    }
    
    return wipe_device(device_path, method, output_file);
}