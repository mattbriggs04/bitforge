spinlock_t lock;
spinlock_init(&lock);
test_reset_atomic();
bool first = spinlock_try_lock(&lock, 1);
bool second = spinlock_try_lock(&lock, 2);
case_passed = first &&
              !second &&
              spinlock_is_locked(&lock) &&
              spinlock_owner(&lock) == 1 &&
              test_cas_calls() == 2;
