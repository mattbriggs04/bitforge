# Debug Frame Length Parser

The starter has bugs. Fix:

```c
int frame_payload_len(const uint8_t *frame, size_t len);
```

Frame format:

```text
byte 0: sync, must be 0xA5
byte 1: payload length high byte
byte 2: payload length low byte
bytes 3..: payload
```

Return the payload length when the frame is valid.
