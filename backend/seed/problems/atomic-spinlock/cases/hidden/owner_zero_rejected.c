spinlock_t lock;
spinlock_init(&lock);
test_reset_atomic();
bool ok = spinlock_try_lock(&lock, 0);
case_passed = !ok &&
              !spinlock_is_locked(&lock) &&
              spinlock_owner(&lock) == 0 &&
              test_cas_calls() == 0;
