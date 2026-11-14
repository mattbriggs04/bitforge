uint32_t word = pack_command(3, 5, true, 120);
uint32_t expected = 3u | (5u << 4) | (1u << 7) | (120u << 8);
case_passed = word == expected;
