unsigned char buf[128];
unsigned char expected[128];
for (size_t i = 0; i < 128; i++) {
    buf[i] = (unsigned char)(i + 1);
    expected[i] = (unsigned char)(i + 1);
}
memmove(expected + 7, expected, 90);
void *ret = bf_memmove(buf + 7, buf, 90);
case_passed = (ret == (buf + 7)) && memcmp(buf, expected, 128) == 0;
