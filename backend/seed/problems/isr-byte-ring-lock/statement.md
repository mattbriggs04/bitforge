# ISR Byte Ring With Critical Sections

Implement a byte ring buffer shared by an ISR producer and a main-loop consumer:

```c
void irq_ring_init(irq_ring_t *rb);
bool irq_ring_push_isr(irq_ring_t *rb, uint8_t byte);
bool irq_ring_pop(irq_ring_t *rb, uint8_t *out);
size_t irq_ring_size(const irq_ring_t *rb);
```

Rules:

- Capacity is `IRQ_RING_CAPACITY`.
- `push_isr` runs in interrupt context and must not call the critical-section hooks.
- `pop` runs in main-loop context and must call `enter_critical`/`exit_critical` exactly once.
- `size` returns a protected snapshot and must also call the hooks exactly once.
- FIFO order must be preserved across wrap-around.
