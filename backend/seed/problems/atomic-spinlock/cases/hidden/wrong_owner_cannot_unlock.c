spinlock_t lock;
spinlock_init(&lock);
bool locked = spinlock_try_lock(&lock, 3);
bool unlocked = spinlock_unlock(&lock, 4);
case_passed = locked &&
              !unlocked &&
              spinlock_is_locked(&lock) &&
              spinlock_owner(&lock) == 3;
