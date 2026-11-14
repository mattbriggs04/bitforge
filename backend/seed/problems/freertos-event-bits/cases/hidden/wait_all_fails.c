event_group_t group;
events_init(&group);
events_set(&group, 0x02u);
bool ready = events_wait(&group, 0x06u, true, true);
case_passed = !ready && events_get(&group) == 0x02u;
