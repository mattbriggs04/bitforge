# Sign Extend Sensor Values

Implement:

```c
int32_t sign_extend(uint32_t value, uint8_t bits);
int16_t saturating_add_i16(int16_t a, int16_t b);
```

`sign_extend` treats the low `bits` bits of `value` as a two's-complement integer.
`saturating_add_i16` clamps to `INT16_MIN` or `INT16_MAX` when the exact sum exceeds 16-bit range.
