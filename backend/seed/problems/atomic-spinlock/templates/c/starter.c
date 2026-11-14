#include <stdbool.h>
#include <stdint.h>

bool atomic_compare_exchange_u8(volatile uint8_t *addr, uint8_t expected, uint8_t desired);

typedef struct {
    volatile uint8_t locked;
    uint32_t owner;
} spinlock_t;

void spinlock_init(spinlock_t *lock) {
    lock->locked = 0;
    lock->owner = 0;
}

bool spinlock_try_lock(spinlock_t *lock, uint32_t owner_id) {
    (void)lock;
    (void)owner_id;
    return false;
}

bool spinlock_unlock(spinlock_t *lock, uint32_t owner_id) {
    (void)lock;
    (void)owner_id;
    return false;
}

bool spinlock_is_locked(const spinlock_t *lock) {
    return lock->locked != 0;
}

uint32_t spinlock_owner(const spinlock_t *lock) {
    return lock->owner;
}
