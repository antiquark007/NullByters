#include "json_log.h"
#include <time.h>
#include <string.h>

static const char* bus_name(bus_t b) {
    switch (b) {
        case BUS_SATA: return "SATA/SAS";
        case BUS_NVME: return "NVMe";
        case BUS_USB:  return "USB";
        default:       return "Unknown";
    }
}

int write_json_cert(FILE *fp,
                    const char *tool_name,
                    const char *version,
                    const device_t *dev,
                    const hpa_dco_report_t *hpa,
                    wipe_mode_t mode,
                    const op_result_t *res,
                    const char *started_iso8601,
                    const char *finished_iso8601)
{
    const char *nist = nist_level_for(bus_name(dev->bus), mode, res->method_code);

    fprintf(fp, "{\n");
    fprintf(fp, "  \"tool\": \"%s\",\n", tool_name);
    fprintf(fp, "  \"version\": \"%s\",\n", version);
    fprintf(fp, "  \"device\": {\n");
    fprintf(fp, "    \"path\": \"%s\",\n", dev->path);
    fprintf(fp, "    \"model\": \"%s\",\n", dev->model);
    fprintf(fp, "    \"serial\": \"%s\",\n", dev->serial);
    fprintf(fp, "    \"firmware\": \"%s\",\n", dev->firmware);
    fprintf(fp, "    \"bus\": \"%s\",\n", bus_name(dev->bus));
    fprintf(fp, "    \"size_bytes\": %llu\n", (unsigned long long)dev->size_bytes);
    fprintf(fp, "  },\n");

    fprintf(fp, "  \"hpa_dco\": {\n");
    fprintf(fp, "    \"hpa_present\": %s,\n", hpa->hpa_present ? "true" : "false");
    fprintf(fp, "    \"hpa_max\": %llu,\n", hpa->hpa_max);
    fprintf(fp, "    \"native_max\": %llu,\n", hpa->native_max);
    fprintf(fp, "    \"dco_limited\": %s,\n", hpa->dco_limited ? "true" : "false");
    fprintf(fp, "    \"suggest\": \"%s\"\n", hpa->suggest_cmd);
    fprintf(fp, "  },\n");

    fprintf(fp, "  \"operation\": {\n");
    fprintf(fp, "    \"mode\": \"%s\",\n", mode == MODE_PURGE ? "purge" : "clear");
    fprintf(fp, "    \"method\": \"%s\",\n", res->method_name);
    fprintf(fp, "    \"nist_level\": \"%s\",\n", nist);
    fprintf(fp, "    \"started_at\": \"%s\",\n", started_iso8601);
    fprintf(fp, "    \"finished_at\": \"%s\",\n", finished_iso8601);
    fprintf(fp, "    \"exit_code\": %d,\n", res->exit_code);
    fprintf(fp, "    \"verify_note\": \"%s\"\n", res->verify_note);
    fprintf(fp, "  },\n");

    // Keep transcript short; escape minimal quotes
    fprintf(fp, "  \"transcript\": \"");
    for (const char *p = res->transcript; *p; ++p) {
        if (*p == '\"') fputc('\\', fp), fputc('\"', fp);
        else if (*p == '\n' || *p == '\r') fputc(' ', fp);
        else fputc(*p, fp);
    }
    fprintf(fp, "\"\n");

    fprintf(fp, "}\n");
    return 0;
}
