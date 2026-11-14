#include <stdbool.h>
#include <stddef.h>
#include <stdint.h>

typedef struct {
    int16_t torque;
    uint16_t speed;
    bool armed;
    bool fault;
    uint8_t counter;
} can_control_t;

int parse_can_control(const uint8_t *payload, size_t len, can_control_t *out) {
    (void)payload;
    (void)len;
    (void)out;
    return -1;
}
