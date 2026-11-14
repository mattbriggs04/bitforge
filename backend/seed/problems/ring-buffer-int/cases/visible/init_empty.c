ring_t rb;
ring_init(&rb);
int out = -1;
bool popped = ring_pop(&rb, &out);
case_passed = (ring_size(&rb) == 0) && (!popped) && (out == -1);
