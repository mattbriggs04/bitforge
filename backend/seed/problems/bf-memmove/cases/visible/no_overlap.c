unsigned char src[] = {1, 2, 3, 4};
unsigned char dst[] = {0, 0, 0, 0};
void *ret = bf_memmove(dst, src, sizeof(src));
case_passed = (ret == dst) && memcmp(dst, src, sizeof(src)) == 0;
