uint8_t raw[40];
pool_init(raw + 1, 32);
void *a = alloc_aligned(1, 8);
case_passed = a != 0 && (((uintptr_t)a & 7u) == 0);
