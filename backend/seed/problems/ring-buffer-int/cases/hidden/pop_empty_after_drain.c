bf_ring_t rb;
bf_ring_init(&rb);
(void)bf_ring_push(&rb, 7);
int out = 0;
(void)bf_ring_pop(&rb, &out);
bool again = bf_ring_pop(&rb, &out);
case_passed = !again && bf_ring_size(&rb) == 0;
