pool_t pool;
pool_init(&pool);
void *ptrs[POOL_BLOCKS];
bool unique = true;
for (int i = 0; i < POOL_BLOCKS; i++) {
    ptrs[i] = pool_alloc(&pool);
    unique = unique && ptrs[i] != 0;
    for (int j = 0; j < i; j++) {
        unique = unique && ptrs[i] != ptrs[j];
    }
}
case_passed = unique && pool_available(&pool) == 0;
