uint8_t payload[] = {0};
can_control_t out = {0};
int a = parse_can_control(0, 8, &out);
int b = parse_can_control(payload, 8, 0);
case_passed = a == -1 && b == -1;
