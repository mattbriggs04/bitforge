# CAS Spinlock With Owner Checks

Implement a small spinlock API:

```c
void spinlock_init(spinlock_t *lock);
bool spinlock_try_lock(spinlock_t *lock, uint32_t owner_id);
bool spinlock_unlock(spinlock_t *lock, uint32_t owner_id);
bool spinlock_is_locked(const spinlock_t *lock);
uint32_t spinlock_owner(const spinlock_t *lock);
```

Rules:

- Owner id `0` means no owner and cannot acquire the lock.
- `try_lock` must use `atomic_compare_exchange_u8` to change `locked` from `0` to `1`.
- A failed `try_lock` must not change the owner.
- Only the current owner can unlock.
- Unlocking clears both `locked` and `owner`.
