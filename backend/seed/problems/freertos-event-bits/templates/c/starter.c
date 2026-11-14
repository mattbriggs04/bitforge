#include <stdbool.h>
#include <stdint.h>

#define EVENT_USER_MASK 0x00FFFFFFu

typedef struct {
    uint32_t bits;
} event_group_t;

void events_init(event_group_t *group) {
    group->bits = 0;
}

void events_set(event_group_t *group, uint32_t bits) {
    (void)group;
    (void)bits;
}

void events_clear(event_group_t *group, uint32_t bits) {
    (void)group;
    (void)bits;
}

bool events_wait(event_group_t *group, uint32_t mask, bool clear_on_exit, bool wait_all) {
    (void)group;
    (void)mask;
    (void)clear_on_exit;
    (void)wait_all;
    return false;
}

uint32_t events_get(const event_group_t *group) {
    return group->bits;
}
