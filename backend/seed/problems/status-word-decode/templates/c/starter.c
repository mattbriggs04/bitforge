#include <stdbool.h>
#include <stdint.h>

typedef struct {
    bool ready;
    bool degraded;
    bool fault;
    uint8_t subsystem;
    uint8_t error_code;
} status_t;

void decode_status(uint16_t word, status_t *out) {
    (void)word;
    out->ready = false;
    out->degraded = false;
    out->fault = false;
    out->subsystem = 0;
    out->error_code = 0;
}

bool status_faulted(uint16_t word) {
    (void)word;
    return false;
}
