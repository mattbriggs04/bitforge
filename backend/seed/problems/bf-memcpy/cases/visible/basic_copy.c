unsigned char src[] = {0xDE, 0xAD, 0xBE, 0xEF};
unsigned char dst[] = {0x00, 0x00, 0x00, 0x00};
void *ret = bf_memcpy(dst, src, sizeof(src));
case_passed = (ret == dst) && (memcmp(dst, src, sizeof(src)) == 0);
