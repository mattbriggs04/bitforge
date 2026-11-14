uint8_t payload[] = {0x01u, 0x00u, 0x00u, 0x01u, 0x03u, 0x09u, 0xAAu, 0xA0u};
can_control_t out = {0};
int rc = parse_can_control(payload, sizeof(payload), &out);
case_passed = rc == 0 && out.armed && out.fault && out.counter == 9;
