irq_ring_t rb;
irq_ring_init(&rb);
for (uint8_t i = 0; i < IRQ_RING_CAPACITY; i++) {
    (void)irq_ring_push_isr(&rb, i);
}
uint8_t out = 0;
for (uint8_t i = 0; i < 3; i++) {
    (void)irq_ring_pop(&rb, &out);
}
(void)irq_ring_push_isr(&rb, 90);
(void)irq_ring_push_isr(&rb, 91);
test_reset_lock_stats();
size_t size = irq_ring_size(&rb);
case_passed = size == 7 &&
              test_lock_enters() == 1 &&
              test_lock_exits() == 1 &&
              test_lock_errors() == 0;
