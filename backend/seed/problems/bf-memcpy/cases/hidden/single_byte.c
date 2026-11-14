unsigned char src[] = {0x42};
unsigned char dst[] = {0x99};
void *ret = memcpy(dst, src, 1);
case_passed = (ret == dst) && dst[0] == 0x42;
