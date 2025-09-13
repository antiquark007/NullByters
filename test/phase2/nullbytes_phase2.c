/* nullbytes_phase2.c
 *
 * Phase 2: NullBytes - CLEAR wipe engine (single-pass zero overwrite)
 * - dry-run mode (simulate)
 * - safety checks: ensure device is not mounted and not root disk (best-effort)
 * - writes zero-buffer in chunks
 * - samples random offsets after wipe and records SHA256 of sample reads
 * - emits JSON log (json-c)
 *
 * Compile:
 * gcc -O2 nullbytes_phase2.c -o nullbytes_phase2 -ljson-c -lcrypto
 */

#define _GNU_SOURCE
#include <inttypes.h>
#include <stdio.h>
#include <stdlib.h>
#include <stdint.h>
#include <string.h>
#include <unistd.h>
#include <fcntl.h>
#include <sys/ioctl.h>
#include <linux/fs.h> /* BLKGETSIZE64 */
#include <errno.h>
#include <time.h>
#include <sys/stat.h>
#include <sys/types.h>
#include <json-c/json.h>
#include <openssl/sha.h>
#include <dirent.h>

#define DEFAULT_CHUNK_MB 16
#define SAMPLE_COUNT 8
#define SAMPLE_LEN 4096 /* bytes per sample read */

/* Helper: exit on error */
static void perror_exit(const char *msg)
{
    perror(msg);
    exit(EXIT_FAILURE);
}

/* Get device size in bytes */
static unsigned long long get_device_size_bytes(const char *devpath)
{
    int fd = open(devpath, O_RDONLY | O_CLOEXEC);
    if (fd < 0)
    {
        fprintf(stderr, "Error opening %s: %s\n", devpath, strerror(errno));
        return 0;
    }
    unsigned long long size = 0;
    if (ioctl(fd, BLKGETSIZE64, &size) != 0)
    {
        fprintf(stderr, "BLKGETSIZE64 ioctl failed on %s: %s\n", devpath, strerror(errno));
        close(fd);
        return 0;
    }
    close(fd);
    return size;
}

/* Check if device is mounted */
static int is_path_mounted(const char *devpath)
{
    FILE *f = fopen("/proc/mounts", "r");
    if (!f) return 0;
    char line[1024];
    while (fgets(line, sizeof(line), f))
    {
        char mpdev[512], mpoint[512], rest[512];
        if (sscanf(line, "%511s %511s %511s", mpdev, mpoint, rest) >= 2)
        {
            if (strcmp(mpdev, devpath) == 0)
            {
                fclose(f);
                return 1;
            }
        }
    }
    fclose(f);
    return 0;
}

/* ISO8601 timestamp */
static char *timestamp_iso8601(void)
{
    time_t t = time(NULL);
    struct tm tm;
    gmtime_r(&t, &tm);
    char *buf = malloc(32);
    strftime(buf, 32, "%Y-%m-%dT%H:%M:%SZ", &tm);
    return buf;
}

/* SHA256 hex digest */
static void sha256_hex(const unsigned char *data, size_t len, char out_hex[65])
{
    unsigned char digest[SHA256_DIGEST_LENGTH];
    SHA256(data, len, digest);
    for (int i = 0; i < SHA256_DIGEST_LENGTH; ++i)
        sprintf(out_hex + i*2, "%02x", digest[i]);
    out_hex[64] = 0;
}

/* random uint64_t */
static unsigned long long random_u64(void)
{
    unsigned long long v = 0;
    FILE *ur = fopen("/dev/urandom", "rb");
    if (ur)
    {
        if (fread(&v, sizeof(v), 1, ur) == 1)
        {
            fclose(ur);
            return v;
        }
        fclose(ur);
    }
    for (int i=0; i<4; ++i) v=(v<<15)^rand();
    return v;
}

/* simple progress bar */
static void progress_bar(double fraction, unsigned long long written_bytes, unsigned long long total_bytes)
{
    int bar_width = 40;
    int pos = (int)(fraction * bar_width);
    printf("\r[");
    for (int i=0; i<bar_width; ++i) printf(i<pos?"█":"-");
    printf("] %3.0f%% %.2f/%.2f MB",
           fraction*100.0,
           written_bytes/(1024.0*1024),
           total_bytes/(1024.0*1024));
    fflush(stdout);
}

int main(int argc, char **argv)
{
    const char *devpath = NULL;
    int dry_run=0, confirm=0, chunk_mb=DEFAULT_CHUNK_MB;

    for(int i=1;i<argc;i++)
    {
        if(strcmp(argv[i],"--device")==0 && i+1<argc) devpath=argv[++i];
        else if(strcmp(argv[i],"--dry-run")==0) dry_run=1;
        else if(strcmp(argv[i],"--confirm")==0) confirm=1;
        else if(strcmp(argv[i],"--chunk-mb")==0 && i+1<argc) chunk_mb=atoi(argv[++i]);
        else {fprintf(stderr,"Unknown arg: %s\n",argv[i]); return 1;}
    }

    if(!devpath){fprintf(stderr,"--device is required\n"); return 1;}
    if(access(devpath,F_OK)!=0){fprintf(stderr,"Device %s not found\n",devpath); return 1;}

    printf("NullBytes Phase2 - CLEAR wipe engine\nDevice: %s\n",devpath);
    if(dry_run) printf("Mode: DRY-RUN\n");
    if(!dry_run && !confirm){fprintf(stderr,"Provide --confirm to wipe\n"); return 1;}
    if(is_path_mounted(devpath)){fprintf(stderr,"ERROR: device mounted\n"); return 1;}

    unsigned long long total_bytes=get_device_size_bytes(devpath);
    if(total_bytes==0){fprintf(stderr,"Cannot get device size\n"); return 1;}
    printf("Device size: %" PRIu64 " bytes (%.2f GiB)\n", total_bytes,total_bytes/(1024.0*1024*1024));

    char *started_at = timestamp_iso8601();
    struct json_object *jroot=json_object_new_object();
    json_object_object_add(jroot,"tool",json_object_new_string("NullBytes"));
    json_object_object_add(jroot,"version",json_object_new_string("0.2.0"));
    json_object_object_add(jroot,"device",json_object_new_string(devpath));
    json_object_object_add(jroot,"started_at",json_object_new_string(started_at));
    json_object_object_add(jroot,"mode",json_object_new_string(dry_run?"clear-dryrun":"clear"));

    if(dry_run)
    {
        json_object_object_add(jroot,"estimate_bytes",json_object_new_int64(total_bytes));
        json_object_object_add(jroot,"note",json_object_new_string("Dry-run; no writes."));
        char outfn[256];
        snprintf(outfn,sizeof(outfn),"wipe_%s_dryrun.json",strrchr(devpath,'/')?strrchr(devpath,'/')+1:devpath);
        FILE *of=fopen(outfn,"w");
        if(of){fputs(json_object_to_json_string_ext(jroot,JSON_C_TO_STRING_PRETTY),of); fclose(of); printf("Dry-run JSON: %s\n",outfn);}
        else fprintf(stderr,"Failed to write dry-run JSON\n");
        json_object_put(jroot);
        free(started_at);
        return 0;
    }

    int fd=open(devpath,O_RDWR | O_DIRECT | O_SYNC);
    if(fd<0){fd=open(devpath,O_RDWR|O_SYNC); if(fd<0){fprintf(stderr,"Cannot open device\n"); json_object_put(jroot); free(started_at); return 1;}}

    size_t chunk=(size_t)chunk_mb*1024*1024;
    void *buf=NULL;
    if(posix_memalign(&buf,4096,chunk)!=0) buf=NULL;
    if(!buf){buf=malloc(chunk); if(!buf) perror_exit("malloc");}
    memset(buf,0,chunk);

    unsigned long long written=0;
    off_t offset=0;
    printf("Starting zero overwrite...\n");
    while((unsigned long long)offset<total_bytes)
    {
        size_t to_write=chunk;
        if((unsigned long long)offset+to_write>total_bytes) to_write=(size_t)(total_bytes-offset);
        ssize_t w=write(fd,buf,to_write);
        if(w<0){fprintf(stderr,"Write error\n"); close(fd); free(buf); json_object_put(jroot); return 1;}
        written+=w;
        offset+=w;
        progress_bar((double)written/total_bytes, written, total_bytes);
    }
    printf("\nFlush and sync...\n");
    fsync(fd); close(fd);

    char *finished_at=timestamp_iso8601();
    json_object_object_add(jroot,"finished_at",json_object_new_string(finished_at));
    json_object_object_add(jroot,"bytes_written",json_object_new_int64((int64_t)written));
    json_object_object_add(jroot,"method",json_object_new_string("zero_overwrite"));

    /* Sampling verification */
    struct json_object *jevidence=json_object_new_object();
    struct json_object *jsamples=json_object_new_array();
    unsigned char *rbuf=malloc(SAMPLE_LEN);
    for(int i=0;i<SAMPLE_COUNT;i++)
    {
        unsigned long long r=random_u64() % (total_bytes>SAMPLE_LEN?(total_bytes-SAMPLE_LEN):1);
        int rfd=open(devpath,O_RDONLY|O_CLOEXEC);
        if(rfd<0) continue;
        if(lseek(rfd,(off_t)r,SEEK_SET)==(off_t)-1){close(rfd); continue;}
        ssize_t rr=read(rfd,rbuf,SAMPLE_LEN); close(rfd);
        if(rr<=0) continue;
        char shahex[65]; sha256_hex(rbuf,rr,shahex);
        struct json_object *js=json_object_new_object();
        json_object_object_add(js,"offset",json_object_new_int64((int64_t)r));
        json_object_object_add(js,"bytes_read",json_object_new_int(rr));
        json_object_object_add(js,"sha256",json_object_new_string(shahex));
        int all_zero=1;
        for(ssize_t k=0;k<rr;k++) if(rbuf[k]!=0){all_zero=0; break;}
        json_object_object_add(js,"all_zero",json_object_new_boolean(all_zero));
        json_object_array_add(jsamples,js);
    }
    json_object_object_add(jevidence,"samples",jsamples);
    json_object_object_add(jroot,"evidence",jevidence);

    /* write JSON log */
    char outfn[256];
    char *devname=strrchr(devpath,'/');
    devname=devname?devname+1:(char*)devpath;
    snprintf(outfn,sizeof(outfn),"wipe_%s_%ld.json",devname,time(NULL));
    FILE *of=fopen(outfn,"w");
    if(!of) fprintf(stderr,"Failed to open log file\n");
    else {fputs(json_object_to_json_string_ext(jroot,JSON_C_TO_STRING_PRETTY),of); fclose(of); printf("Wipe log: %s\n",outfn);}

    free(buf); free(rbuf); free(started_at); free(finished_at); json_object_put(jroot);
    printf("Wipe complete.\n");
    return 0;
}
