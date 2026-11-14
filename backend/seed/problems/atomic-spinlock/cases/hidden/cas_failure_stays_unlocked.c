spinlock_t lock;
spinlock_init(&lock);
test_reset_atomic();
test_force_cas_failure(1);
bool ok = spinlock_try_lock(&lock, 9);
case_passed = !ok &&
              !spinlock_is_locked(&lock) &&
              spinlock_owner(&lock) == 0 &&
              test_cas_calls() == 1;
