spinlock_t lock;
spinlock_init(&lock);
bool ok = spinlock_unlock(&lock, 1);
case_passed = !ok &&
              !spinlock_is_locked(&lock) &&
              spinlock_owner(&lock) == 0;
