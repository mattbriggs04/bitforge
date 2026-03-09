debounce_button_t db;
debounce_init(&db, 0);
uint8_t seq[] = {
    0,1,0,1,1,1,1,1,1,1,1,
    1,0,1,0,0,0,0,0,0,0,0
};
int rises = 0;
int falls = 0;
for (size_t i = 0; i < sizeof(seq); i++) {
    debounce_timer_isr(&db, seq[i]);
    rises += debounce_take_rising_edge(&db);
    falls += debounce_take_falling_edge(&db);
}
case_passed = rises == 1 && falls == 1 && db.stable_state == 0;
