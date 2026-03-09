unsigned char src[33];
unsigned char dst[33];
for (size_t i = 0; i < 33; i++) {
    src[i] = (unsigned char)((i * 13) ^ 0x5A);
    dst[i] = 0xEE;
}
void *ret = bf_memcpy(dst, src, 33);
case_passed = (ret == dst) && memcmp(dst, src, 33) == 0;
