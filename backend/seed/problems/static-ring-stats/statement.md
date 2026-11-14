# Track Static Sensor Stats

Implement a module-level rolling statistics API:

```c
void stats_reset(void);
void stats_record(uint16_t sample);
uint8_t stats_count(void);
uint16_t stats_min(void);
uint16_t stats_max(void);
uint16_t stats_avg(void);
```

Store only the last 8 samples. `avg` is integer division over the active window.
