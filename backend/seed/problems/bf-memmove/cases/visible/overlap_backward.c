unsigned char buf[] = {10, 11, 12, 13, 14, 15, 16, 17};
unsigned char expected[] = {10, 11, 12, 13, 14, 15, 16, 17};
memmove(expected, expected + 3, 4);
void *ret = bf_memmove(buf, buf + 3, 4);
case_passed = (ret == buf) && memcmp(buf, expected, sizeof(buf)) == 0;
