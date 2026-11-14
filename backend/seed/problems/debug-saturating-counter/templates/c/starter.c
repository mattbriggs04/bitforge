#include <stdint.h>

static uint8_t count;
static uint8_t max_count;

void counter_reset(uint8_t max_value) {
    count = 0;
    if (max_count == 0) {
        max_count = max_value;
    }
}

void counter_inc(void) {
    if (count <= max_count) {
        count++;
    }
}

void counter_dec(void) {
    count--;
}

uint8_t counter_get(void) {
    return count;
}
