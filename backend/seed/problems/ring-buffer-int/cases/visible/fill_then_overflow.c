ring_t rb;
ring_init(&rb);
bool ok = true;
for (int i = 0; i < RING_CAPACITY; i++) {
    ok = ok && ring_push(&rb, i + 1);
}
bool overflow = ring_push(&rb, 99);
case_passed = ok && !overflow && ring_size(&rb) == RING_CAPACITY;
