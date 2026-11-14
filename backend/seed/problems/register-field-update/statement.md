# Update a Volatile Register Field

Implement:

```c
uint32_t reg_update_field(volatile uint32_t *reg, uint32_t mask, uint8_t shift, uint32_t value);
```

The function updates `*reg` with a read-modify-write operation and returns the previous register value.

The new field value is `(value << shift) & mask`; all unmasked bits are preserved.
