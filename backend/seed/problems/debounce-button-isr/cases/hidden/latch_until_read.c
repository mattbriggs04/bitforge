debounce_button_t db;
debounce_init(&db, 0);
for (int i = 0; i < 8; i++) {
    debounce_timer_isr(&db, 1);
}
uint8_t first = debounce_take_rising_edge(&db);
uint8_t second = debounce_take_rising_edge(&db);
case_passed = (first == 1) && (second == 0);
