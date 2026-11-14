#include <stdbool.h>
#include <stdint.h>

typedef struct {
    uint8_t actuator_id;
    uint8_t mode;
    bool enabled;
    int16_t setpoint;
} command_t;

uint32_t pack_command(uint8_t actuator_id, uint8_t mode, bool enabled, int16_t setpoint) {
    (void)actuator_id;
    (void)mode;
    (void)enabled;
    (void)setpoint;
    return 0;
}

void unpack_command(uint32_t word, command_t *out) {
    (void)word;
    out->actuator_id = 0;
    out->mode = 0;
    out->enabled = false;
    out->setpoint = 0;
}
