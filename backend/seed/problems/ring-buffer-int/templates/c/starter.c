#include <stdbool.h>
#include <stddef.h>

#define RING_CAPACITY 8

typedef struct {
    int data[RING_CAPACITY];
    size_t head;
    size_t tail;
    size_t count;
} ring_t;

void ring_init(ring_t *rb) {
    rb->head = 0;
    rb->tail = 0;
    rb->count = 0;
}

bool ring_push(ring_t *rb, int value) {
    (void)rb;
    (void)value;
    return false;
}

bool ring_pop(ring_t *rb, int *out) {
    (void)rb;
    (void)out;
    return false;
}

size_t ring_size(const ring_t *rb) {
    return rb->count;
}
