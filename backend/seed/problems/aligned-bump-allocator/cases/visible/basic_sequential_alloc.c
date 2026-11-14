uint8_t pool[32];
pool_init(pool, sizeof(pool));
void *a = alloc_aligned(8, 1);
void *b = alloc_aligned(4, 1);
case_passed = a == pool && b == pool + 8 && pool_remaining() == 20;
