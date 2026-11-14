pool_t pool;
pool_init(&pool);
void *a = pool_alloc(&pool);
void *b = pool_alloc(&pool);
pool_free(&pool, b);
void *c = pool_alloc(&pool);
case_passed = a != 0 && b != 0 && c == b && pool_available(&pool) == POOL_BLOCKS - 2;
