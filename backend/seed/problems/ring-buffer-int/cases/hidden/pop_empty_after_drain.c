ring_t rb;
ring_init(&rb);
(void)ring_push(&rb, 7);
int out = 0;
(void)ring_pop(&rb, &out);
bool again = ring_pop(&rb, &out);
case_passed = !again && ring_size(&rb) == 0;
