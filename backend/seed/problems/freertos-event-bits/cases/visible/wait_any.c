event_group_t group;
events_init(&group);
events_set(&group, 0x04u);
bool ready = events_wait(&group, 0x06u, false, false);
case_passed = ready && events_get(&group) == 0x04u;
