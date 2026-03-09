bf_ring_t rb;
bf_ring_init(&rb);
bool ok = true;
for (int i = 0; i < BF_RING_CAPACITY; i++) {
    ok = ok && bf_ring_push(&rb, i + 1);
}
bool overflow = bf_ring_push(&rb, 99);
case_passed = ok && !overflow && bf_ring_size(&rb) == BF_RING_CAPACITY;
