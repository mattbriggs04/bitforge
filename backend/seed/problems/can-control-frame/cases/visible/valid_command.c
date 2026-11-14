uint8_t payload[] = {0x34u, 0x12u, 0x01u, 0xF4u, 0x01u, 0x7Au, 0x00u, 0xA8u};
can_control_t out = {0};
int rc = parse_can_control(payload, sizeof(payload), &out);
case_passed = rc == 0 &&
              out.torque == 0x1234 &&
              out.speed == 500 &&
              out.armed &&
              !out.fault &&
              out.counter == 0x7A;
