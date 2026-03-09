unsigned char buf[] = {1, 3, 5, 7, 9, 11, 13};
unsigned char expected[] = {1, 3, 5, 7, 9, 11, 13};
memmove(expected + 1, expected, 5);
void *ret = bf_memmove(buf + 1, buf, 5);
case_passed = (ret == (buf + 1)) && memcmp(buf, expected, sizeof(buf)) == 0;
