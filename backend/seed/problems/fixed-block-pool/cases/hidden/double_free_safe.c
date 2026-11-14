pool_t pool;
pool_init(&pool);
void *block = pool_alloc(&pool);
pool_free(&pool, block);
pool_free(&pool, block);
case_passed = pool_available(&pool) == POOL_BLOCKS;
