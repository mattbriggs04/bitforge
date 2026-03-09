unsigned char src[] = {1, 2, 3, 4, 5, 6};
unsigned char dst[] = {9, 9, 9, 9, 9, 9};
void *ret = bf_memcpy(dst, src, 3);
case_passed = (ret == dst) && dst[0] == 1 && dst[1] == 2 && dst[2] == 3 && dst[3] == 9 && dst[4] == 9 && dst[5] == 9;
