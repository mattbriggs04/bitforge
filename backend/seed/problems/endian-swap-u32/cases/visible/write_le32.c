uint8_t out[] = {0, 0, 0, 0};
write_le32(out, 0x12345678u);
case_passed = out[0] == 0x78u && out[1] == 0x56u && out[2] == 0x34u && out[3] == 0x12u;
