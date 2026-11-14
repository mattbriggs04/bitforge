#include <stddef.h>
#include <stdint.h>

#define POOL_BLOCKS 4
#define POOL_BLOCK_SIZE 16

typedef struct {
    uint8_t storage[POOL_BLOCKS][POOL_BLOCK_SIZE];
    uint8_t used[POOL_BLOCKS];
} pool_t;

void pool_init(pool_t *pool) {
    (void)pool;
}

void *pool_alloc(pool_t *pool) {
    (void)pool;
    return 0;
}

void pool_free(pool_t *pool, void *ptr) {
    (void)pool;
    (void)ptr;
}

uint8_t pool_available(const pool_t *pool) {
    (void)pool;
    return 0;
}
