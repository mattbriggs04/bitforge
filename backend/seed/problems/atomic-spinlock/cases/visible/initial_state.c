spinlock_t lock;
spinlock_init(&lock);
case_passed = !spinlock_is_locked(&lock) && spinlock_owner(&lock) == 0;
