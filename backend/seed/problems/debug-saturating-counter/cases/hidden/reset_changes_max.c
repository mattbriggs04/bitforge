counter_reset(2);
counter_inc();
counter_inc();
counter_inc();
uint8_t first = counter_get();
counter_reset(5);
for (int i = 0; i < 6; i++) {
    counter_inc();
}
case_passed = first == 2 && counter_get() == 5;
