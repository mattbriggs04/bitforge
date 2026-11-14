volatile uint32_t reg = 0xFFFF0000u;
uint32_t old = reg_update_field(&reg, 0x0000000Fu, 0u, 5u);
case_passed = old == 0xFFFF0000u && reg == 0xFFFF0005u;
