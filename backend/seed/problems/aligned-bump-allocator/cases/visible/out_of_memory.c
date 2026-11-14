uint8_t pool[8];
pool_init(pool, sizeof(pool));
void *a = alloc_aligned(16, 1);
case_passed = a == 0 && pool_remaining() == 8;
