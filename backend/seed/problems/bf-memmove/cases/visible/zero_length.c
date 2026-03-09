unsigned char buf[] = {9, 8, 7, 6};
unsigned char before[] = {9, 8, 7, 6};
void *ret = bf_memmove(buf + 1, buf, 0);
case_passed = (ret == (buf + 1)) && memcmp(buf, before, sizeof(buf)) == 0;
