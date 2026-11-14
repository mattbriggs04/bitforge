#include <stdbool.h>
#include <stdint.h>

typedef struct {
    uint32_t bitmap;
} slots_t;

void slots_init(slots_t *slots) {
    slots->bitmap = 0u;
}

int slots_alloc(slots_t *slots) {
    (void)slots;
    return -1;
}

void slots_free(slots_t *slots, int slot) {
    (void)slots;
    (void)slot;
}

bool slots_is_used(const slots_t *slots, int slot) {
    (void)slots;
    (void)slot;
    return false;
}

uint8_t slots_used_count(const slots_t *slots) {
    (void)slots;
    return 0;
}
