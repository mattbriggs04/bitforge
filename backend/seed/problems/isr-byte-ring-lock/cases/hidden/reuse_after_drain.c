irq_ring_t rb;
irq_ring_init(&rb);
for (uint8_t i = 0; i < IRQ_RING_CAPACITY; i++) {
    (void)irq_ring_push_isr(&rb, (uint8_t)(i + 1));
}
uint8_t out = 0;
for (uint8_t i = 0; i < IRQ_RING_CAPACITY; i++) {
    (void)irq_ring_pop(&rb, &out);
}
bool pushed = irq_ring_push_isr(&rb, 0x5A);
bool popped = irq_ring_pop(&rb, &out);
case_passed = pushed && popped && out == 0x5A && irq_ring_size(&rb) == 0;
