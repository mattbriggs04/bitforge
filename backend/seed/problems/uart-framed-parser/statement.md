# UART Framed Parser

Implement a streaming UART parser:

```c
void uart_init(uart_parser_t *parser);
bool uart_feed(uart_parser_t *parser, uint8_t byte, uart_frame_t *out);
```

Frame format:

- Sync byte: `0x7E`.
- Length byte: `1..4`.
- Payload: `length` bytes.
- Checksum: XOR of `length` and all payload bytes.

`uart_feed` returns `true` only when a complete valid frame is written to `out`.
