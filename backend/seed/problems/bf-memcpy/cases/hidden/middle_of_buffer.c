unsigned char src[] = {10, 20, 30, 40};
unsigned char dst[] = {1, 2, 3, 4, 5, 6, 7, 8};
void *ret = bf_memcpy(dst + 2, src, sizeof(src));
case_passed = (ret == (dst + 2)) && dst[0] == 1 && dst[1] == 2 && dst[2] == 10 && dst[3] == 20 && dst[4] == 30 && dst[5] == 40 && dst[6] == 7 && dst[7] == 8;
