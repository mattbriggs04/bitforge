#include <stdbool.h>
#include <stddef.h>
#include <stdint.h>

struct spi_sample {
    uint8_t channel;
    bool fault;
    uint16_t value;
};

int parse_spi_sensor(const uint8_t *rx, size_t len, struct spi_sample *out) {
    (void)rx;
    (void)len;
    (void)out;
    return 0;
}
