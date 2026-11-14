# FreeRTOS-Style Event Bits

Implement a small event-bit group model:

```c
void events_init(event_group_t *group);
void events_set(event_group_t *group, uint32_t bits);
void events_clear(event_group_t *group, uint32_t bits);
bool events_wait(event_group_t *group, uint32_t mask, bool clear_on_exit, bool wait_all);
uint32_t events_get(const event_group_t *group);
```

Rules:

- Only bits `0..23` are user event bits.
- `wait_all=false` succeeds if any requested bit is set.
- `wait_all=true` succeeds only if all requested bits are set.
- If a wait succeeds and `clear_on_exit=true`, clear `mask` from the group.
- Failed waits do not change the group.
