debounce_button_t db;
debounce_init(&db, 0);
int rises = 0;
for (int i = 0; i < 20; i++) {
    debounce_timer_isr(&db, 1);
    rises += debounce_take_rising_edge(&db);
}
case_passed = db.stable_state == 1 && rises == 1 && debounce_take_falling_edge(&db) == 0;
