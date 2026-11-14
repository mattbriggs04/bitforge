spinlock_t lock;
spinlock_init(&lock);
test_reset_atomic();
test_force_cas_failure(1);
bool first = spinlock_try_lock(&lock, 5);
test_force_cas_failure(0);
bool second = spinlock_try_lock(&lock, 5);
case_passed = !first &&
              second &&
              spinlock_is_locked(&lock) &&
              spinlock_owner(&lock) == 5 &&
              test_cas_calls() == 2;
