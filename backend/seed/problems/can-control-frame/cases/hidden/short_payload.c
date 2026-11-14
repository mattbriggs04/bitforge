uint8_t payload[] = {0};
can_control_t out = {0};
int rc = parse_can_control(payload, 7, &out);
case_passed = rc == -2;
