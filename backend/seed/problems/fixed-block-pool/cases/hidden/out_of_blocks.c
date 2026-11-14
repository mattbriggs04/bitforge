pool_t pool;
pool_init(&pool);
for (int i = 0; i < POOL_BLOCKS; i++) {
    (void)pool_alloc(&pool);
}
void *extra = pool_alloc(&pool);
case_passed = extra == 0 && pool_available(&pool) == 0;
