ring_t rb;
ring_init(&rb);
int out = 0;
bool ok = true;
for (int i = 0; i < 20; i++) {
    ok = ok && ring_push(&rb, i);
    ok = ok && ring_pop(&rb, &out) && out == i;
}
case_passed = ok && ring_size(&rb) == 0;
