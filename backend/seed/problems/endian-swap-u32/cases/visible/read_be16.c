uint8_t bytes[] = {0x12u, 0x34u};
case_passed = read_be16(bytes) == 0x1234u;
