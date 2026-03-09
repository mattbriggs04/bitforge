bf_ring_t rb;
bf_ring_init(&rb);
int out = -1;
bool popped = bf_ring_pop(&rb, &out);
case_passed = (bf_ring_size(&rb) == 0) && (!popped) && (out == -1);
