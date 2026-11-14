#include <stdint.h>

void stats_reset(void) {
}

void stats_record(uint16_t sample) {
    (void)sample;
}

uint8_t stats_count(void) {
    return 0;
}

uint16_t stats_min(void) {
    return 0;
}

uint16_t stats_max(void) {
    return 0;
}

uint16_t stats_avg(void) {
    return 0;
}
