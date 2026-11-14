event_group_t group;
events_init(&group);
events_set(&group, 0xFF000001u);
events_clear(&group, 0x80000000u);
case_passed = events_get(&group) == 0x01u;
