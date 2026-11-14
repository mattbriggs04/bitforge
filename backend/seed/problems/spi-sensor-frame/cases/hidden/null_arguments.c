uint8_t rx[] = {0xA5u, 0x00u, 0x12u, 0x34u, 0x83u};
struct spi_sample out = {0, false, 0};
case_passed = parse_spi_sensor(0, sizeof(rx), &out) == -1 &&
              parse_spi_sensor(rx, sizeof(rx), 0) == -1;
