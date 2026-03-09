bf_ring_t rb;
bf_ring_init(&rb);
for (int i = 0; i < BF_RING_CAPACITY; i++) {
    (void)bf_ring_push(&rb, i + 100);
}
int out = 0;
for (int i = 0; i < BF_RING_CAPACITY; i++) {
    (void)bf_ring_pop(&rb, &out);
}
bool push_ok = bf_ring_push(&rb, 77);
bool pop_ok = bf_ring_pop(&rb, &out);
case_passed = push_ok && pop_ok && out == 77 && bf_ring_size(&rb) == 0;
