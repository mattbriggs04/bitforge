debounce_button_t db;
debounce_init(&db, 1);
uint8_t seq[] = {1, 0, 1, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0};
int falls = 0;
for (size_t i = 0; i < sizeof(seq); i++) {
    debounce_timer_isr(&db, seq[i]);
    falls += debounce_take_falling_edge(&db);
}
case_passed = db.stable_state == 0 && falls == 1 && debounce_take_rising_edge(&db) == 0;
