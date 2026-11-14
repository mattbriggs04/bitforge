# Aligned Bump Allocator

Implement:

```c
void pool_init(void *buffer, size_t size);
void *alloc_aligned(size_t size, size_t align);
size_t pool_remaining(void);
```

The allocator returns aligned blocks from the initialized buffer. It never frees individual blocks.
