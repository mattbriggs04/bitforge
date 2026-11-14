# Bitmap Slot Allocator

Implement a fixed 32-slot allocator backed by one `uint32_t` bitmap:

```c
void slots_init(slots_t *slots);
int slots_alloc(slots_t *slots);
void slots_free(slots_t *slots, int slot);
bool slots_is_used(const slots_t *slots, int slot);
uint8_t slots_used_count(const slots_t *slots);
```

Rules:

- Slot ids are `0..31`.
- `alloc` returns the lowest free slot, or `-1` when full.
- `free` ignores invalid slot ids.
- `is_used` returns `false` for invalid slot ids.
