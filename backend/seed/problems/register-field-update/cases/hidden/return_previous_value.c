volatile uint32_t reg = 0xA5A55A5Au;
uint32_t old = reg_update_field(&reg, 0x00FF0000u, 16u, 0x12u);
case_passed = old == 0xA5A55A5Au && reg == 0xA5125A5Au;
