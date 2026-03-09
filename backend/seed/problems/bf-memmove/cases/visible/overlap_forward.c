unsigned char buf[] = {1, 2, 3, 4, 5, 6, 7, 8};
unsigned char expected[] = {1, 2, 3, 4, 5, 6, 7, 8};
memmove(expected + 2, expected, 5);
void *ret = bf_memmove(buf + 2, buf, 5);
case_passed = (ret == (buf + 2)) && memcmp(buf, expected, sizeof(buf)) == 0;
