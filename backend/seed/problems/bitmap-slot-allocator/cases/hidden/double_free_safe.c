slots_t slots;
slots_init(&slots);
(void)slots_alloc(&slots);
slots_free(&slots, 0);
slots_free(&slots, 0);
int next = slots_alloc(&slots);
case_passed = next == 0 && slots_used_count(&slots) == 1;
