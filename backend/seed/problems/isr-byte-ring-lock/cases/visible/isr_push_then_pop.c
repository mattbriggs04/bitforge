irq_ring_t rb;
irq_ring_init(&rb);
test_reset_lock_stats();
bool a = irq_ring_push_isr(&rb, 0x11);
bool b = irq_ring_push_isr(&rb, 0x22);
bool push_unlocked = test_lock_enters() == 0 && test_lock_exits() == 0;
uint8_t out1 = 0;
uint8_t out2 = 0;
bool p1 = irq_ring_pop(&rb, &out1);
bool p2 = irq_ring_pop(&rb, &out2);
case_passed = a && b && push_unlocked && p1 && p2 &&
              out1 == 0x11 && out2 == 0x22 &&
              test_lock_enters() == 2 &&
              test_lock_exits() == 2 &&
              test_lock_depth() == 0 &&
              test_lock_errors() == 0;
