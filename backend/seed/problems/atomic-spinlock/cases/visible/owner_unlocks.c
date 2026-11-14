spinlock_t lock;
spinlock_init(&lock);
bool locked = spinlock_try_lock(&lock, 7);
bool unlocked = spinlock_unlock(&lock, 7);
bool relocked = spinlock_try_lock(&lock, 8);
case_passed = locked &&
              unlocked &&
              relocked &&
              spinlock_is_locked(&lock) &&
              spinlock_owner(&lock) == 8;
