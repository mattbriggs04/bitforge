# Pack Actuator Command Word

Implement packing and unpacking for a compact actuator command:

```c
uint32_t pack_command(uint8_t actuator_id, uint8_t mode, bool enabled, int16_t setpoint);
void unpack_command(uint32_t word, command_t *out);
```

Layout:

- Bits `0..3`: actuator id.
- Bits `4..6`: mode.
- Bit `7`: enabled.
- Bits `8..19`: signed 12-bit setpoint.
- Bits `20..31`: reserved, always zero when packing.

Clamp setpoint to `[-2048, 2047]` before packing.
