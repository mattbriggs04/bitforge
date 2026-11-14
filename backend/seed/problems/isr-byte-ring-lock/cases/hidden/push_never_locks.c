irq_ring_t rb;
irq_ring_init(&rb);
test_reset_lock_stats();
bool ok = true;
for (uint8_t i = 0; i < 5; i++) {
    ok = ok && irq_ring_push_isr(&rb, (uint8_t)(40 + i));
}
case_passed = ok &&
              test_lock_enters() == 0 &&
              test_lock_exits() == 0 &&
              test_lock_depth() == 0 &&
              test_lock_errors() == 0;
