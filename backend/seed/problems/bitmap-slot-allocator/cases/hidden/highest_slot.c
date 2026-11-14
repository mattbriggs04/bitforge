slots_t slots;
slots_init(&slots);
for (int i = 0; i < 31; i++) {
    (void)slots_alloc(&slots);
}
int high = slots_alloc(&slots);
case_passed = high == 31 &&
              slots_is_used(&slots, 31) &&
              slots_used_count(&slots) == 32;
