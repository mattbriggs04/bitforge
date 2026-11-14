slots_t slots;
slots_init(&slots);
(void)slots_alloc(&slots);
(void)slots_alloc(&slots);
(void)slots_alloc(&slots);
slots_free(&slots, 1);
int reused = slots_alloc(&slots);
case_passed = reused == 1 && slots_used_count(&slots) == 3;
