# Decode Telemetry Status Word

Decode a 16-bit telemetry status word:

```c
void decode_status(uint16_t word, status_t *out);
bool status_faulted(uint16_t word);
```

Layout:

- Bit `0`: ready.
- Bit `1`: degraded.
- Bit `2`: fault.
- Bits `4..7`: subsystem id.
- Bits `8..13`: error code.
- Other bits are reserved.

`status_faulted` returns `true` only when the fault bit is set.
