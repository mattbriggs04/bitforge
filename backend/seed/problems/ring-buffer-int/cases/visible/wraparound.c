bf_ring_t rb;
bf_ring_init(&rb);
for (int i = 0; i < 6; i++) {
    (void)bf_ring_push(&rb, i);
}
int drop = 0;
(void)bf_ring_pop(&rb, &drop);
(void)bf_ring_pop(&rb, &drop);
(void)bf_ring_pop(&rb, &drop);
(void)bf_ring_push(&rb, 100);
(void)bf_ring_push(&rb, 101);
(void)bf_ring_push(&rb, 102);

int expected[] = {3, 4, 5, 100, 101, 102};
int out = 0;
bool ok = true;
for (size_t i = 0; i < sizeof(expected) / sizeof(expected[0]); i++) {
    ok = ok && bf_ring_pop(&rb, &out) && out == expected[i];
}
case_passed = ok && bf_ring_size(&rb) == 0;
