unsigned char src[] = {0x00, 0x10, 0x00, 0x7F, 0x00};
unsigned char dst[] = {0xFF, 0xFF, 0xFF, 0xFF, 0xFF};
void *ret = bf_memcpy(dst, src, sizeof(src));
case_passed = (ret == dst) && (memcmp(dst, src, sizeof(src)) == 0);
