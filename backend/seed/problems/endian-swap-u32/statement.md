# Read Big-Endian Fields

Implement:

```c
uint16_t read_be16(const uint8_t *p);
uint32_t read_be32(const uint8_t *p);
void write_le32(uint8_t *p, uint32_t value);
```

Use explicit byte operations. Do not cast `uint8_t *` to wider integer pointers.
