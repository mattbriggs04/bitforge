uint32_t word = pack_command(4, 3, true, -1);
case_passed = (word & 0xFFF00000u) == 0u;
