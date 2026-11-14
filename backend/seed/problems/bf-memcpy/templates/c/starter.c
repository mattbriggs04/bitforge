#include <stddef.h>

void *memcpy(void *dest, const void *src, size_t n) {
    (void)src;
    (void)n;
    return dest;
}
