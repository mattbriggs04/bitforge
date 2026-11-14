irq_ring_t rb;
irq_ring_init(&rb);
uint8_t out = 0xAA;
test_reset_lock_stats();
bool popped = irq_ring_pop(&rb, &out);
case_passed = !popped &&
              out == 0xAA &&
              test_lock_enters() == 1 &&
              test_lock_exits() == 1 &&
              test_lock_depth() == 0 &&
              test_lock_errors() == 0;
