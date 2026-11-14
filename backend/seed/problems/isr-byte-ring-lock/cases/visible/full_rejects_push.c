irq_ring_t rb;
irq_ring_init(&rb);
bool ok = true;
for (uint8_t i = 0; i < IRQ_RING_CAPACITY; i++) {
    ok = ok && irq_ring_push_isr(&rb, (uint8_t)(i + 1));
}
bool overflow = irq_ring_push_isr(&rb, 99);
test_reset_lock_stats();
size_t size = irq_ring_size(&rb);
case_passed = ok && !overflow && size == IRQ_RING_CAPACITY &&
              test_lock_enters() == 1 &&
              test_lock_exits() == 1 &&
              test_lock_errors() == 0;
