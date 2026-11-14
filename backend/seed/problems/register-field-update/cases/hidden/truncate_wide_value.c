volatile uint32_t reg = 0x00000000u;
reg_update_field(&reg, 0x00000F00u, 8u, 0x3Fu);
case_passed = reg == 0x00000F00u;
