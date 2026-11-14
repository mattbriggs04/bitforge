uint8_t bytes[] = {0xFFu, 0xFFu, 0xFFu, 0xFFu};
uint8_t out[] = {0, 0, 0, 0};
write_le32(out, 0xFFFFFFFFu);
case_passed = read_be32(bytes) == 0xFFFFFFFFu &&
              out[0] == 0xFFu && out[1] == 0xFFu && out[2] == 0xFFu && out[3] == 0xFFu;
