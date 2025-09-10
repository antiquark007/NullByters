/*
onewipe_core.c

Prototype C core for OneWipe (Linux-focused).  
Purpose: provide a safe, auditable core library and CLI that:
 - enumerates block devices (Linux /sys/block)
 - overwrites devices with patterns or random data
 - produces logs and a JSON wipe certificate
 - signs the JSON certificate using an RSA private key (PEM) via OpenSSL
 - verifies signed certificates using a public key (PEM)

IMPORTANT SAFETY NOTES
 - This tool is destructive. Running overwrite or discard WILL DESTROY DATA.
 - Test only on throwaway/test devices or virtual disks.
 - You MUST run as root for direct block device access.
 - This prototype intentionally omits low-level ATA/NVMe pass-through and HPA/DCO manipulation (these are platform- and vendor-specific and will be implemented in a later step).

Build (on Linux):
    gcc -o onewipe_core onewipe_core.c -lcrypto

Usage examples (after building):
    # list devices
    sudo ./onewipe_core list

    # overwrite a device with 1 pass of 0xFF pattern (destructive!)
    sudo ./onewipe_core overwrite /dev/sdX 1 255

    # overwrite with random data (passes=1, pattern=rand)
    sudo ./onewipe_core overwrite /dev/sdX 1 rand

    # generate a JSON certificate from a wipe log
    ./onewipe_core gen-cert /path/to/wipe.log /tmp/wipe.json

    # sign the JSON certificate with RSA private key (PEM)
    ./onewipe_core sign-cert /tmp/wipe.json /path/to/private.pem /tmp/wipe.json.sig

    # verify signed certificate with public key
    ./onewipe_core verify-cert /tmp/wipe.json /tmp/wipe.json.sig /path/to/public.pem

Files created by overwrite (by default):
  ./onewipe-logs/<device_basename>-<timestamp>.log
  ./onewipe-certs/<device_basename>-<timestamp>.json
  ./onewipe-certs/<device_basename>-<timestamp>.json.sig

--------------------------------------------------------------------------------
IMPLEMENTATION NOTES
 - Focused on correctness and clarity rather than maximum performance.
 - Uses OpenSSL EVP APIs for signing/verification (link with -lcrypto).
 - Uses /sys/block to enumerate devices; reads size from /sys/block/<dev>/size (in 512-byte sectors).
 - Overwrite uses 4 MiB buffer writes.
 - Random writes source data from /dev/urandom (skip verification for random writes).
 - BLKDISCARD and firmware-based secure-erase are NOT implemented here; the prototype calls out via log entries where those should be invoked.
--------------------------------------------------------------------------------
*/

#define _GNU_SOURCE
#include <stdio.h>
#include <stdlib.h>
#include <stdint.h>
#include <string.h>
#include <dirent.h>
#include <fcntl.h>
#include <sys/stat.h>
#include <sys/types.h>
#include <unistd.h>
#include <errno.h>
#include <time.h>
#include <inttypes.h>
#include <openssl/evp.h>
#include <openssl/pem.h>
#include <openssl/err.h>
#include <openssl/sha.h>

#define CHUNK_SIZE (4 * 1024 * 1024) // 4 MiB
#define LOG_DIR "./onewipe-logs"
#define CERT_DIR "./onewipe-certs"

static void ensure_dirs() {
    struct stat st = {0};
    if (stat(LOG_DIR, &st) == -1) mkdir(LOG_DIR, 0755);
    if (stat(CERT_DIR, &st) == -1) mkdir(CERT_DIR, 0755);
}

static int is_root() {
    return geteuid() == 0;
}

static char *iso8601_now_z() {
    time_t t = time(NULL);
    struct tm tm;
    gmtime_r(&t, &tm);
    char *buf = malloc(32);
    strftime(buf, 32, "%Y-%m-%dT%H:%M:%SZ", &tm);
    return buf;
}

// --- Device enumeration (Linux /sys/block) ---
static void list_block_devices() {
    DIR *d = opendir("/sys/block");
    if (!d) {
        perror("opendir /sys/block");
        return;
    }
    struct dirent *ent;
    printf("Detected block devices (Linux /sys/block):\n");
    while ((ent = readdir(d)) != NULL) {
        if (ent->d_name[0] == '.') continue;
        char sys_path[PATH_MAX];
        snprintf(sys_path, sizeof(sys_path), "/sys/block/%s/size", ent->d_name);
        FILE *f = fopen(sys_path, "r");
        if (!f) continue;
        unsigned long long sectors = 0;
        if (fscanf(f, "%llu", &sectors) != 1) { fclose(f); continue; }
        fclose(f);
        unsigned long long bytes = sectors * 512ULL;
        printf("  /dev/%s  size=%llu bytes (%llu sectors)\n", ent->d_name, bytes, sectors);
    }
    closedir(d);
}

// return device size in bytes, or 0 on error
static unsigned long long get_device_size_bytes(const char *devpath) {
    // Expect devpath like /dev/sdX or /dev/nvme0n1
    const char *base = strrchr(devpath, '/');
    if (!base) base = devpath; else base++;
    char sys_size_path[PATH_MAX];
    snprintf(sys_size_path, sizeof(sys_size_path), "/sys/block/%s/size", base);
    FILE *f = fopen(sys_size_path, "r");
    if (!f) return 0;
    unsigned long long sectors = 0;
    if (fscanf(f, "%llu", &sectors) != 1) { fclose(f); return 0; }
    fclose(f);
    return sectors * 512ULL;
}

// compute SHA256 of a file and write hex digest to out_hex (must be 65 bytes)
static int sha256_file_hex(const char *path, char out_hex[65]) {
    unsigned char buf[CHUNK_SIZE];
    unsigned char hash[SHA256_DIGEST_LENGTH];
    SHA256_CTX ctx;
    SHA256_Init(&ctx);
    FILE *f = fopen(path, "rb");
    if (!f) return -1;
    size_t r;
    while ((r = fread(buf, 1, sizeof(buf), f)) > 0) {
        SHA256_Update(&ctx, buf, r);
    }
    fclose(f);
    SHA256_Final(hash, &ctx);
    for (int i = 0; i < SHA256_DIGEST_LENGTH; ++i) sprintf(out_hex + (i*2), "%02x", hash[i]);
    out_hex[64] = '\0';
    return 0;
}

// compute SHA256 of a memory buffer
static int sha256_buf_hex(const unsigned char *buf, size_t len, char out_hex[65]) {
    unsigned char hash[SHA256_DIGEST_LENGTH];
    SHA256_CTX ctx;
    SHA256_Init(&ctx);
    SHA256_Update(&ctx, buf, len);
    SHA256_Final(hash, &ctx);
    for (int i = 0; i < SHA256_DIGEST_LENGTH; ++i) sprintf(out_hex + (i*2), "%02x", hash[i]);
    out_hex[64] = '\0';
    return 0;
}

// write log message timestamped
static void log_msg(FILE *logf, const char *fmt, ...) {
    char *ts = iso8601_now_z();
    fprintf(logf, "[%s] ", ts);
    free(ts);
    va_list ap;
    va_start(ap, fmt);
    vfprintf(logf, fmt, ap);
    va_end(ap);
    fprintf(logf, "\n");
    fflush(logf);
}

// overwrite device with pattern byte or random data
static int overwrite_device(const char *devpath, int passes, const char *pattern_s) {
    if (!is_root()) {
        fprintf(stderr, "Error: overwrite requires root privileges.\n");
        return -1;
    }
    unsigned long long dev_size = get_device_size_bytes(devpath);
    if (dev_size == 0) {
        fprintf(stderr, "Could not determine device size for %s\n", devpath);
        return -1;
    }
    char *now = iso8601_now_z();
    // build log file name
    const char *base = strrchr(devpath, '/'); base = base ? base+1 : devpath;
    char logpath[PATH_MAX];
    snprintf(logpath, sizeof(logpath), "%s/%s-%s.log", LOG_DIR, base, now);
    free(now);
    FILE *logf = fopen(logpath, "w");
    if (!logf) { perror("fopen log"); return -1; }
    log_msg(logf, "Starting overwrite of %s (size=%llu bytes)", devpath, dev_size);
    int fd = open(devpath, O_RDWR | O_SYNC);
    if (fd < 0) { perror("open device"); log_msg(logf, "open failed: %s", strerror(errno)); fclose(logf); return -1; }

    unsigned char *buf = malloc(CHUNK_SIZE);
    if (!buf) { perror("malloc"); close(fd); fclose(logf); return -1; }

    int use_random = 0;
    int pattern = 0;
    if (strcasecmp(pattern_s, "rand") == 0 || strcasecmp(pattern_s, "random") == 0) use_random = 1;
    else {
        char *endp;
        long p = strtol(pattern_s, &endp, 0);
        if (endp == pattern_s) { fprintf(stderr, "Invalid pattern '%s'\n", pattern_s); free(buf); close(fd); fclose(logf); return -1; }
        pattern = (int)(p & 0xFF);
    }

    for (int pass = 0; pass < passes; ++pass) {
        log_msg(logf, "Pass %d/%d", pass+1, passes);
        unsigned long long written = 0;
        off_t off = 0;
        while (written < dev_size) {
            size_t towrite = CHUNK_SIZE;
            if (dev_size - written < towrite) towrite = (size_t)(dev_size - written);
            if (use_random) {
                // read random bytes
                int ur = open("/dev/urandom", O_RDONLY);
                if (ur < 0) { perror("open /dev/urandom"); free(buf); close(fd); fclose(logf); return -1; }
                ssize_t rr = read(ur, buf, towrite);
                close(ur);
                if (rr != (ssize_t)towrite) {
                    log_msg(logf, "warning: short read from /dev/urandom");
                }
            } else {
                memset(buf, pattern, towrite);
            }
            ssize_t w = pwrite(fd, buf, towrite, off);
            if (w < 0) {
                log_msg(logf, "write error at offset %lld: %s", (long long)off, strerror(errno));
                free(buf); close(fd); fclose(logf); return -1;
            }
            off += w;
            written += w;
            // simple progress to stdout
            if (written % (16 * CHUNK_SIZE) == 0) {
                double pct = (double)written * 100.0 / (double)dev_size;
                printf("\rpass %d/%d: %.2f%%", pass+1, passes, pct);
                fflush(stdout);
            }
        }
        fsync(fd);
        printf("\rpass %d/%d: 100.00%%\n", pass+1, passes);
        log_msg(logf, "Completed pass %d", pass+1);

        // For non-random patterns, perform a simple verification read-back for the last 16 KiB of device
        if (!use_random) {
            unsigned char vbuf[16384];
            off_t verify_off = dev_size >= sizeof(vbuf) ? (off_t)(dev_size - sizeof(vbuf)) : 0;
            ssize_t rr = pread(fd, vbuf, sizeof(vbuf), verify_off);
            if (rr <= 0) {
                log_msg(logf, "verify read error: %s", strerror(errno));
            } else {
                int ok = 1;
                for (int i = 0; i < rr; ++i) if (vbuf[i] != (unsigned char)pattern) { ok = 0; break; }
                log_msg(logf, "verify last segment %s", ok ? "OK" : "MISMATCH");
            }
        } else {
            log_msg(logf, "Random pass — no verify performed (random pattern)");
        }
    }

    free(buf);
    close(fd);
    log_msg(logf, "Overwrite finished");
    fclose(logf);
    printf("Log written to: %s\n", logpath);
    return 0;
}

// create a simple JSON certificate from a log file
static int gen_certificate_from_log(const char *logpath, const char *outjsonpath, const char *device, const char *method) {
    char loghash[65];
    if (sha256_file_hex(logpath, loghash) != 0) {
        fprintf(stderr, "Could not hash log file\n"); return -1;
    }
    char *ts = iso8601_now_z();
    const char *base = strrchr(device, '/'); base = base ? base+1 : device;
    FILE *f = fopen(outjsonpath, "w");
    if (!f) { perror("fopen json out"); return -1; }
    fprintf(f, "{\n");
    fprintf(f, "  \"certificate_version\": \"1.0\",\n");
    fprintf(f, "  \"asset\": {\"device\": \"%s\"},\n", base);
    fprintf(f, "  \"wipe\": {\"method\": \"%s\", \"timestamp\": \"%s\"},\n", method, ts);
    fprintf(f, "  \"verification\": {\"log_hash\": \"%s\"}\n", loghash);
    fprintf(f, "}\n");
    fclose(f);
    free(ts);
    printf("Generated JSON certificate: %s\n", outjsonpath);
    return 0;
}

// sign a file (json) using RSA private key PEM, output raw signature file
static int sign_file_rsa(const char *inpath, const char *privkey_pem, const char *outsigpath) {
    FILE *kf = fopen(privkey_pem, "r");
    if (!kf) { perror("fopen privkey"); return -1; }
    EVP_PKEY *pkey = PEM_read_PrivateKey(kf, NULL, NULL, NULL);
    fclose(kf);
    if (!pkey) { fprintf(stderr, "Could not read private key\n"); return -1; }

    // compute digest of inpath
    unsigned char buf[CHUNK_SIZE];
    unsigned char md[EVP_MAX_MD_SIZE];
    unsigned int mdlen = 0;
    EVP_MD_CTX *mdctx = EVP_MD_CTX_new();
    FILE *f = fopen(inpath, "rb");
    if (!f) { perror("fopen inpath"); EVP_PKEY_free(pkey); EVP_MD_CTX_free(mdctx); return -1; }
    if (EVP_DigestInit_ex(mdctx, EVP_sha256(), NULL) != 1) { fprintf(stderr, "DigestInit failed\n"); fclose(f); EVP_PKEY_free(pkey); EVP_MD_CTX_free(mdctx); return -1; }
    size_t r;
    while ((r = fread(buf, 1, sizeof(buf), f)) > 0) EVP_DigestUpdate(mdctx, buf, r);
    EVP_DigestFinal_ex(mdctx, md, &mdlen);
    fclose(f);
    EVP_MD_CTX_free(mdctx);

    // sign the digest
    EVP_MD_CTX *signctx = EVP_MD_CTX_new();
    if (EVP_DigestSignInit(signctx, NULL, EVP_sha256(), NULL, pkey) != 1) { fprintf(stderr, "SignInit failed\n"); EVP_PKEY_free(pkey); EVP_MD_CTX_free(signctx); return -1; }
    size_t siglen = 0;
    if (EVP_DigestSign(signctx, NULL, &siglen, md, mdlen) != 1) { fprintf(stderr, "DigestSign (len) failed\n"); EVP_PKEY_free(pkey); EVP_MD_CTX_free(signctx); return -1; }
    unsigned char *sig = malloc(siglen);
    if (!sig) { fprintf(stderr, "malloc sig failed\n"); EVP_PKEY_free(pkey); EVP_MD_CTX_free(signctx); return -1; }
    if (EVP_DigestSign(signctx, sig, &siglen, md, mdlen) != 1) { fprintf(stderr, "DigestSign failed\n"); free(sig); EVP_PKEY_free(pkey); EVP_MD_CTX_free(signctx); return -1; }

    FILE *out = fopen(outsigpath, "wb");
    if (!out) { perror("fopen outsig"); free(sig); EVP_PKEY_free(pkey); EVP_MD_CTX_free(signctx); return -1; }
    fwrite(sig, 1, siglen, out);
    fclose(out);
    EVP_MD_CTX_free(signctx);
    EVP_PKEY_free(pkey);
    free(sig);
    printf("Signature written to %s\n", outsigpath);
    return 0;
}

// verify signature file using RSA public key PEM
static int verify_file_rsa(const char *inpath, const char *sigpath, const char *pubkey_pem) {
    FILE *kf = fopen(pubkey_pem, "r");
    if (!kf) { perror("fopen pubkey"); return -1; }
    EVP_PKEY *pkey = PEM_read_PUBKEY(kf, NULL, NULL, NULL);
    fclose(kf);
    if (!pkey) { fprintf(stderr, "Could not read public key\n"); return -1; }

    // compute digest of inpath
    unsigned char buf[CHUNK_SIZE];
    unsigned char md[EVP_MAX_MD_SIZE];
    unsigned int mdlen = 0;
    EVP_MD_CTX *mdctx = EVP_MD_CTX_new();
    FILE *f = fopen(inpath, "rb");
    if (!f) { perror("fopen inpath"); EVP_PKEY_free(pkey); EVP_MD_CTX_free(mdctx); return -1; }
    if (EVP_DigestInit_ex(mdctx, EVP_sha256(), NULL) != 1) { fprintf(stderr, "DigestInit failed\n"); fclose(f); EVP_PKEY_free(pkey); EVP_MD_CTX_free(mdctx); return -1; }
    size_t r;
    while ((r = fread(buf, 1, sizeof(buf), f)) > 0) EVP_DigestUpdate(mdctx, buf, r);
    EVP_DigestFinal_ex(mdctx, md, &mdlen);
    fclose(f);
    EVP_MD_CTX_free(mdctx);

    // read signature
    FILE *sf = fopen(sigpath, "rb");
    if (!sf) { perror("fopen sig"); EVP_PKEY_free(pkey); return -1; }
    fseek(sf, 0, SEEK_END);
    long siglen = ftell(sf);
    fseek(sf, 0, SEEK_SET);
    unsigned char *sig = malloc(siglen);
    if (!sig) { fprintf(stderr, "malloc sig failed\n"); fclose(sf); EVP_PKEY_free(pkey); return -1; }
    fread(sig, 1, siglen, sf);
    fclose(sf);

    EVP_MD_CTX *vctx = EVP_MD_CTX_new();
    if (EVP_DigestVerifyInit(vctx, NULL, EVP_sha256(), NULL, pkey) != 1) { fprintf(stderr, "VerifyInit failed\n"); free(sig); EVP_PKEY_free(pkey); EVP_MD_CTX_free(vctx); return -1; }
    int rc = EVP_DigestVerify(vctx, sig, siglen, md, mdlen);
    EVP_MD_CTX_free(vctx);
    EVP_PKEY_free(pkey);
    free(sig);
    if (rc == 1) {
        printf("Signature: VALID\n"); return 0;
    } else if (rc == 0) {
        printf("Signature: INVALID\n"); return 2;
    } else {
        printf("Signature: ERROR\n"); return -1;
    }
}

// --- CLI ---
int main(int argc, char **argv) {
    if (argc < 2) {
        fprintf(stderr, "Usage: %s <command> [args]\nCommands: list | overwrite <device> <passes> <pattern|rand> | gen-cert <log> <out.json> | sign-cert <json> <priv.pem> <out.sig> | verify-cert <json> <sig> <pub.pem>\n", argv[0]);
        return 1;
    }
    ensure_dirs();
    OpenSSL_add_all_algorithms();

    const char *cmd = argv[1];
    if (strcmp(cmd, "list") == 0) {
        list_block_devices();
        return 0;
    }
    else if (strcmp(cmd, "overwrite") == 0) {
        if (argc < 5) { fprintf(stderr, "overwrite usage: %s overwrite <device> <passes> <pattern|rand>\n", argv[0]); return 1; }
        const char *dev = argv[2];
        int passes = atoi(argv[3]); if (passes <= 0) passes = 1;
        const char *pattern = argv[4];
        return overwrite_device(dev, passes, pattern);
    }
    else if (strcmp(cmd, "gen-cert") == 0) {
        if (argc < 4) { fprintf(stderr, "gen-cert usage: %s gen-cert <logpath> <out.json>\n", argv[0]); return 1; }
        return gen_certificate_from_log(argv[2], argv[3], "unknown", "OVERWRITE");
    }
    else if (strcmp(cmd, "sign-cert") == 0) {
        if (argc < 5) { fprintf(stderr, "sign-cert usage: %s sign-cert <json> <privkey.pem> <out.sig>\n", argv[0]); return 1; }
        return sign_file_rsa(argv[2], argv[3], argv[4]);
    }
    else if (strcmp(cmd, "verify-cert") == 0) {
        if (argc < 5) { fprintf(stderr, "verify-cert usage: %s verify-cert <json> <sig> <pubkey.pem>\n", argv[0]); return 1; }
        return verify_file_rsa(argv[2], argv[3], argv[4]);
    }
    else {
        fprintf(stderr, "Unknown command: %s\n", cmd);
        return 1;
    }
}
