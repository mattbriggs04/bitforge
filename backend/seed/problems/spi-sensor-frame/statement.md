# Parse an SPI Sensor Frame

Implement:

```c
struct spi_sample {
    uint8_t channel;
    bool fault;
    uint16_t value;
};

int parse_spi_sensor(const uint8_t *rx, size_t len, struct spi_sample *out);
```

Frame format:

```text
byte 0: sync, must be 0xA5
byte 1: status, bit 7 = fault, bits 2..0 = channel
byte 2: value high byte
byte 3: value low byte
byte 4: checksum = byte0 ^ byte1 ^ byte2 ^ byte3
```

Return `0` on success, otherwise a negative error code.
