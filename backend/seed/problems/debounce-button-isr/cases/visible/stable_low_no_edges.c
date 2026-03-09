debounce_button_t db;
debounce_init(&db, 0);
for (int i = 0; i < 16; i++) {
    debounce_timer_isr(&db, 0);
}
case_passed = db.stable_state == 0 &&
    debounce_take_rising_edge(&db) == 0 &&
    debounce_take_falling_edge(&db) == 0;
