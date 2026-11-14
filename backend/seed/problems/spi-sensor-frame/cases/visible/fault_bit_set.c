uint8_t rx[] = {0xA5u, 0x85u, 0x00u, 0x2Au, 0x0Au};
struct spi_sample out = {0, false, 0};
int rc = parse_spi_sensor(rx, sizeof(rx), &out);
case_passed = rc == 0 && out.channel == 5 && out.fault == true && out.value == 0x002Au;
