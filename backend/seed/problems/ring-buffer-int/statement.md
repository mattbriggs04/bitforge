Implement a fixed-size integer ring buffer.

You are given a buffer type and function signatures:

- `void bf_ring_init(bf_ring_t *rb)`
- `bool bf_ring_push(bf_ring_t *rb, int value)`
- `bool bf_ring_pop(bf_ring_t *rb, int *out)`
- `size_t bf_ring_size(const bf_ring_t *rb)`

Behavior:
- Capacity is `BF_RING_CAPACITY`.
- `push` returns `false` if the buffer is full.
- `pop` returns `false` if the buffer is empty.
- Data order must be FIFO.

This mirrors common firmware queueing logic used between ISR and main-loop contexts.
