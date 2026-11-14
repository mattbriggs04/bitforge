slots_t slots;
slots_init(&slots);
bool all_free = true;
for (int i = 0; i < 32; i++) {
    all_free = all_free && !slots_is_used(&slots, i);
}
case_passed = all_free && slots_used_count(&slots) == 0;
