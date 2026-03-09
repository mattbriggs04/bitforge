bf_ring_t rb;
bf_ring_init(&rb);
for (int i = 0; i < BF_RING_CAPACITY; i++) {
    (void)bf_ring_push(&rb, i + 1);
}
int out = 0;
for (int i = 0; i < 5; i++) {
    (void)bf_ring_pop(&rb, &out);
}
for (int i = 0; i < 5; i++) {
    (void)bf_ring_push(&rb, 50 + i);
}
int expected[] = {6, 7, 8, 50, 51, 52, 53, 54};
bool ok = true;
for (size_t i = 0; i < sizeof(expected) / sizeof(expected[0]); i++) {
    ok = ok && bf_ring_pop(&rb, &out) && out == expected[i];
}
case_passed = ok && bf_ring_size(&rb) == 0;
