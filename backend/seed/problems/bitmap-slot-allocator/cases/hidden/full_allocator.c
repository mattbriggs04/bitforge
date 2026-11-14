slots_t slots;
slots_init(&slots);
bool ok = true;
for (int i = 0; i < 32; i++) {
    ok = ok && slots_alloc(&slots) == i;
}
int extra = slots_alloc(&slots);
case_passed = ok && extra == -1 && slots_used_count(&slots) == 32;
