uint8_t payload[] = {0x85u, 0xFFu, 0x00u, 0x64u, 0x00u, 0x02u, 0x00u, 0x1Cu};
can_control_t out = {0};
int rc = parse_can_control(payload, sizeof(payload), &out);
case_passed = rc == 0 && out.torque == -123 && out.speed == 100;
