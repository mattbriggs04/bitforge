unsigned char buf[128];
unsigned char expected[128];
for (size_t i = 0; i < 128; i++) {
    buf[i] = (unsigned char)(255 - i);
    expected[i] = (unsigned char)(255 - i);
}
memmove(expected, expected + 11, 100);
void *ret = bf_memmove(buf, buf + 11, 100);
case_passed = (ret == buf) && memcmp(buf, expected, 128) == 0;
