irq_ring_t rb;
irq_ring_init(&rb);
for (uint8_t i = 0; i < 6; i++) {
    (void)irq_ring_push_isr(&rb, i);
}
uint8_t drop = 0;
(void)irq_ring_pop(&rb, &drop);
(void)irq_ring_pop(&rb, &drop);
(void)irq_ring_pop(&rb, &drop);
(void)irq_ring_push_isr(&rb, 100);
(void)irq_ring_push_isr(&rb, 101);
(void)irq_ring_push_isr(&rb, 102);

uint8_t expected[] = {3, 4, 5, 100, 101, 102};
bool ok = true;
uint8_t out = 0;
for (size_t i = 0; i < sizeof(expected) / sizeof(expected[0]); i++) {
    ok = ok && irq_ring_pop(&rb, &out) && out == expected[i];
}
case_passed = ok && irq_ring_size(&rb) == 0 && test_lock_errors() == 0;
