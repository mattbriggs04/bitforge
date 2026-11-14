ring_t rb;
ring_init(&rb);
for (int i = 0; i < RING_CAPACITY; i++) {
    (void)ring_push(&rb, i + 100);
}
int out = 0;
for (int i = 0; i < RING_CAPACITY; i++) {
    (void)ring_pop(&rb, &out);
}
bool push_ok = ring_push(&rb, 77);
bool pop_ok = ring_pop(&rb, &out);
case_passed = push_ok && pop_ok && out == 77 && ring_size(&rb) == 0;
