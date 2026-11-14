# Fixed Block Memory Pool

Implement a fixed-size block pool:

```c
void pool_init(pool_t *pool);
void *pool_alloc(pool_t *pool);
void pool_free(pool_t *pool, void *ptr);
uint8_t pool_available(const pool_t *pool);
```

The pool contains `POOL_BLOCKS` blocks. Each block has `POOL_BLOCK_SIZE` bytes.

Rules:

- `alloc` returns `NULL` when no blocks are available.
- `free` ignores invalid pointers.
- A valid pointer must be exactly the start of a block from the same pool.
- Double free must be harmless.
