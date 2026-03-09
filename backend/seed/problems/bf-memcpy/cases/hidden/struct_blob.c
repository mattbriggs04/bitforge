struct Payload {
    unsigned short a;
    unsigned int b;
    unsigned char c[3];
};
struct Payload src = {0x1122, 0xAABBCCDD, {9, 8, 7}};
struct Payload dst = {0};
void *ret = bf_memcpy(&dst, &src, sizeof(src));
case_passed = (ret == &dst) && memcmp(&dst, &src, sizeof(src)) == 0;
