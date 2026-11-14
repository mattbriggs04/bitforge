#include <stddef.h>

void pool_init(void *buffer, size_t size) {
    (void)buffer;
    (void)size;
}

void *alloc_aligned(size_t size, size_t align) {
    (void)size;
    (void)align;
    return 0;
}

size_t pool_remaining(void) {
    return 0;
}
