uint8_t rx[] = {0xA5u, 0x01u, 0x00u, 0x10u, 0x00u};
struct spi_sample out = {7, true, 0xEEEEu};
int rc = parse_spi_sensor(rx, sizeof(rx), &out);
case_passed = rc == -4 && out.channel == 7 && out.fault == true && out.value == 0xEEEEu;
