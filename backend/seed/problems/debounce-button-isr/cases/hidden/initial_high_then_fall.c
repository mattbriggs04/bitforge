debounce_button_t db;
debounce_init(&db, 1);
for (int i = 0; i < 8; i++) {
    debounce_timer_isr(&db, 0);
}
case_passed = db.stable_state == 0 && debounce_take_falling_edge(&db) == 1 && debounce_take_falling_edge(&db) == 0;
