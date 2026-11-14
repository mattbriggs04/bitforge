slots_t slots;
slots_init(&slots);
(void)slots_alloc(&slots);
(void)slots_alloc(&slots);
slots_free(&slots, -1);
slots_free(&slots, 32);
case_passed = slots_is_used(&slots, 0) &&
              slots_is_used(&slots, 1) &&
              !slots_is_used(&slots, -1) &&
              !slots_is_used(&slots, 32) &&
              slots_used_count(&slots) == 2;
