# MMIO Status Snapshot

Implement a helper for a memory-mapped status register:

```c
int read_status(volatile uint32_t *status_reg,
                   volatile uint32_t *clear_reg,
                   status_snapshot_t *out);
```

Register layout:

- Bit `0`: ready.
- Bit `1`: fault, write-one-to-clear.
- Bit `2`: overrun, write-one-to-clear.
- Bits `4..6`: mode.
- Bits `8..15`: error code.

Return `0` on success and `-1` for invalid pointers.

On success, write `fault | overrun` bits to `*clear_reg`. If `clear_reg` is `NULL`, still decode the snapshot but skip the write.
