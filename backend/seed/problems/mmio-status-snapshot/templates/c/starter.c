#include <stdbool.h>
#include <stdint.h>

typedef struct {
    bool ready;
    bool fault;
    bool overrun;
    uint8_t mode;
    uint8_t error_code;
} status_snapshot_t;

int read_status(volatile uint32_t *status_reg,
                   volatile uint32_t *clear_reg,
                   status_snapshot_t *out) {
    (void)status_reg;
    (void)clear_reg;
    (void)out;
    return -1;
}
