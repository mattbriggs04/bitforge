#include <stdbool.h>
#include <stddef.h>

#define BF_RING_CAPACITY 8

typedef struct {
    int data[BF_RING_CAPACITY];
    size_t head;
    size_t tail;
    size_t count;
} bf_ring_t;

void bf_ring_init(bf_ring_t *rb) {
    rb->head = 0;
    rb->tail = 0;
    rb->count = 0;
}

bool bf_ring_push(bf_ring_t *rb, int value) {
    (void)rb;
    (void)value;
    return false;
}

bool bf_ring_pop(bf_ring_t *rb, int *out) {
    (void)rb;
    (void)out;
    return false;
}

size_t bf_ring_size(const bf_ring_t *rb) {
    return rb->count;
}
