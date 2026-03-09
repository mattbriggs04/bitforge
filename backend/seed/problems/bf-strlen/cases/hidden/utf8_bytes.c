const unsigned char input[] = {0xC3, 0xA9, 0x00, 0x41};
size_t got = bf_strlen((const char *)input);
size_t expected = 2;
case_passed = (got == expected);
