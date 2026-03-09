unsigned char src[] = {0xAA, 0xBB, 0xCC};
unsigned char dst[] = {0x11, 0x22, 0x33};
unsigned char before[] = {0x11, 0x22, 0x33};
void *ret = bf_memcpy(dst, src, 0);
case_passed = (ret == dst) && (memcmp(dst, before, sizeof(dst)) == 0);
