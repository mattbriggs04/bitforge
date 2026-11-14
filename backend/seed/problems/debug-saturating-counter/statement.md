# Debug Saturating Counter

The starter has bugs. Fix this small static-state API:

```c
void counter_reset(uint8_t max_value);
void counter_inc(void);
void counter_dec(void);
uint8_t counter_get(void);
```

The counter starts at `0`, saturates at `max_value`, and never underflows below `0`.
