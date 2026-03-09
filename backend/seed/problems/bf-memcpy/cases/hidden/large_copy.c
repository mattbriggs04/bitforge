unsigned char src[512];
unsigned char dst[512];
for (size_t i = 0; i < 512; i++) {
    src[i] = (unsigned char)(i & 0xFF);
    dst[i] = 0;
}
void *ret = bf_memcpy(dst, src, 512);
case_passed = (ret == dst) && memcmp(dst, src, 512) == 0;
