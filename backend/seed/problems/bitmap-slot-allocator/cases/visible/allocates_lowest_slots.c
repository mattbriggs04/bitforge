slots_t slots;
slots_init(&slots);
int a = slots_alloc(&slots);
int b = slots_alloc(&slots);
int c = slots_alloc(&slots);
case_passed = a == 0 && b == 1 && c == 2 &&
              slots_is_used(&slots, 0) &&
              slots_is_used(&slots, 1) &&
              slots_is_used(&slots, 2) &&
              slots_used_count(&slots) == 3;
