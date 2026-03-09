unsigned char buf[] = {42, 43, 44, 45, 46, 47, 48, 49};
unsigned char expected[] = {42, 43, 44, 45, 46, 47, 48, 49};
memmove(expected + 2, expected + 4, 3);
void *ret = bf_memmove(buf + 2, buf + 4, 3);
case_passed = (ret == (buf + 2)) && memcmp(buf, expected, sizeof(buf)) == 0;
