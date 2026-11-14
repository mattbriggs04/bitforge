uint8_t pool[16];
pool_init(pool, sizeof(pool));
void *a = alloc_aligned(0, 4);
case_passed = a == 0 && pool_remaining() == 16;
