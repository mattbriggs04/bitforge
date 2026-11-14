irq_ring_t rb;
irq_ring_init(&rb);
for (uint8_t i = 0; i < IRQ_RING_CAPACITY; i++) {
    (void)irq_ring_push_isr(&rb, (uint8_t)(10 + i));
}
test_reset_lock_stats();
bool ok = true;
uint8_t out = 0;
for (uint8_t i = 0; i < IRQ_RING_CAPACITY; i++) {
    ok = ok && irq_ring_pop(&rb, &out) && out == (uint8_t)(10 + i);
}
case_passed = ok &&
              test_lock_enters() == IRQ_RING_CAPACITY &&
              test_lock_exits() == IRQ_RING_CAPACITY &&
              test_lock_depth() == 0 &&
              test_lock_errors() == 0;
