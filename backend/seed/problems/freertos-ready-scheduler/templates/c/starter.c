#include <stdbool.h>
#include <stdint.h>

void rtos_init(void) {
}

int task_create(uint8_t id, uint8_t priority) {
    (void)id;
    (void)priority;
    return -1;
}

void task_set_ready(uint8_t id, bool ready) {
    (void)id;
    (void)ready;
}

void task_delay(uint8_t id, uint32_t ticks) {
    (void)id;
    (void)ticks;
}

void rtos_tick(void) {
}

int schedule_next(void) {
    return -1;
}
