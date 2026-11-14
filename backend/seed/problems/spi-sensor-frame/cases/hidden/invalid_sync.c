uint8_t rx[] = {0x5Au, 0x00u, 0x12u, 0x34u, 0x7Cu};
struct spi_sample out = {1, true, 2};
int rc = parse_spi_sensor(rx, sizeof(rx), &out);
case_passed = rc == -3 && out.channel == 1 && out.fault == true && out.value == 2;
