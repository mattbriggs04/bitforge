pool_t pool;
pool_init(&pool);
void *block = pool_alloc(&pool);
uint8_t stack_byte = 0;
pool_free(&pool, &stack_byte);
pool_free(&pool, ((uint8_t *)block) + 1);
case_passed = pool_available(&pool) == POOL_BLOCKS - 1;
