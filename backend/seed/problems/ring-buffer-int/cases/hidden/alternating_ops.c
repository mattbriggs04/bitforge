bf_ring_t rb;
bf_ring_init(&rb);
int out = 0;
bool ok = true;
for (int i = 0; i < 20; i++) {
    ok = ok && bf_ring_push(&rb, i);
    ok = ok && bf_ring_pop(&rb, &out) && out == i;
}
case_passed = ok && bf_ring_size(&rb) == 0;
