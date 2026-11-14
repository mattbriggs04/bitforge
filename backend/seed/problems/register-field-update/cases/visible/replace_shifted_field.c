volatile uint32_t reg = 0x12345678u;
reg_update_field(&reg, 0x000000F0u, 4u, 0xAu);
case_passed = reg == 0x123456A8u;
