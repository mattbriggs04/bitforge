ring_t rb;
ring_init(&rb);
bool a = ring_push(&rb, 10);
bool b = ring_push(&rb, 20);
int out1 = 0;
int out2 = 0;
bool p1 = ring_pop(&rb, &out1);
bool p2 = ring_pop(&rb, &out2);
case_passed = a && b && p1 && p2 && out1 == 10 && out2 == 20 && ring_size(&rb) == 0;
