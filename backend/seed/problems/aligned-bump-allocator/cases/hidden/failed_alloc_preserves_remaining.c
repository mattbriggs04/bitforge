uint8_t pool[24];
pool_init(pool, sizeof(pool));
void *a = alloc_aligned(8, 4);
void *b = alloc_aligned(32, 4);
void *c = alloc_aligned(16, 4);
case_passed = a == pool && b == 0 && c == pool + 8 && pool_remaining() == 0;
