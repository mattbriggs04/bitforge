#include <stdbool.h>
#include <stddef.h>
#include <stdint.h>

#define IRQ_RING_CAPACITY 8

void enter_critical(void);
void exit_critical(void);

typedef struct {
    uint8_t data[IRQ_RING_CAPACITY];
    volatile size_t head;
    volatile size_t tail;
    volatile size_t count;
} irq_ring_t;

void irq_ring_init(irq_ring_t *rb) {
    rb->head = 0;
    rb->tail = 0;
    rb->count = 0;
}

bool irq_ring_push_isr(irq_ring_t *rb, uint8_t byte) {
    (void)rb;
    (void)byte;
    return false;
}

bool irq_ring_pop(irq_ring_t *rb, uint8_t *out) {
    (void)rb;
    (void)out;
    return false;
}

size_t irq_ring_size(const irq_ring_t *rb) {
    (void)rb;
    return 0;
}
