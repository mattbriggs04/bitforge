pool_t pool;
pool_init(&pool);
case_passed = pool_available(&pool) == POOL_BLOCKS;
