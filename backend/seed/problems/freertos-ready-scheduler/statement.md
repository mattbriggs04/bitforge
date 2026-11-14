# FreeRTOS-Style Ready Scheduler

Implement a small scheduler model:

```c
void rtos_init(void);
int task_create(uint8_t id, uint8_t priority);
void task_set_ready(uint8_t id, bool ready);
void task_delay(uint8_t id, uint32_t ticks);
void rtos_tick(void);
int schedule_next(void);
```

Rules:

- Task ids are `0..7`.
- Higher numeric priority wins.
- Ties choose the lowest task id.
- Created tasks start ready.
- Delayed tasks are not ready until their delay reaches zero.
- `schedule_next` returns `-1` when no task is ready.
