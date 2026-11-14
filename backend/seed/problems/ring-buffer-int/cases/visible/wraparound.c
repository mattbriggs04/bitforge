ring_t rb;
ring_init(&rb);
for (int i = 0; i < 6; i++) {
    (void)ring_push(&rb, i);
}
int drop = 0;
(void)ring_pop(&rb, &drop);
(void)ring_pop(&rb, &drop);
(void)ring_pop(&rb, &drop);
(void)ring_push(&rb, 100);
(void)ring_push(&rb, 101);
(void)ring_push(&rb, 102);

int expected[] = {3, 4, 5, 100, 101, 102};
int out = 0;
bool ok = true;
for (size_t i = 0; i < sizeof(expected) / sizeof(expected[0]); i++) {
    ok = ok && ring_pop(&rb, &out) && out == expected[i];
}
case_passed = ok && ring_size(&rb) == 0;
