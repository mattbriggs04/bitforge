Implement a fixed-size integer ring buffer.

You are given a buffer type and function signatures:

- `void ring_init(ring_t *rb)`
- `bool ring_push(ring_t *rb, int value)`
- `bool ring_pop(ring_t *rb, int *out)`
- `size_t ring_size(const ring_t *rb)`

Behavior:
- Capacity is `RING_CAPACITY`.
- `push` returns `false` if the buffer is full.
- `pop` returns `false` if the buffer is empty.
- Data order must be FIFO.

This mirrors common firmware queueing logic used between ISR and main-loop contexts.
