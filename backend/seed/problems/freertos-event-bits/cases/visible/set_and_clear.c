event_group_t group;
events_init(&group);
events_set(&group, 0x03u);
events_clear(&group, 0x01u);
case_passed = events_get(&group) == 0x02u;
