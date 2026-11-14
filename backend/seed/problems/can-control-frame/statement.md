# Decode CAN Control Frame

Decode an 8-byte CAN control payload:

```c
int parse_can_control(const uint8_t *payload, size_t len, can_control_t *out);
```

Payload layout:

- Bytes `0..1`: torque command, signed int16, little-endian.
- Bytes `2..3`: speed command, unsigned uint16, big-endian.
- Byte `4`: flags, bit 0 = armed, bit 1 = fault.
- Byte `5`: rolling counter.
- Byte `6`: reserved.
- Byte `7`: checksum, XOR of bytes `0..6`.

Return `0` on success, `-1` for null arguments, `-2` for short payload, and `-3` for checksum mismatch.
