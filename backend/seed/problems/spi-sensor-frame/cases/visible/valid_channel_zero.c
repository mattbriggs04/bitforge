uint8_t rx[] = {0xA5u, 0x00u, 0x12u, 0x34u, 0x83u};
struct spi_sample out = {9, true, 0xFFFFu};
int rc = parse_spi_sensor(rx, sizeof(rx), &out);
case_passed = rc == 0 && out.channel == 0 && out.fault == false && out.value == 0x1234u;
