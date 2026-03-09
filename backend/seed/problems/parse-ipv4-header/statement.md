Implement `int parse_ipv4_header(const uint8_t *packet, size_t len, ipv4_header_t *out)`.

Parse an IPv4 header from raw packet bytes and fill the output struct.

Return codes:
- `0` success
- `-1` invalid arguments (`packet == NULL` or `out == NULL`)
- `-2` packet shorter than 20 bytes
- `-3` version is not IPv4
- `-4` IHL is invalid (`ihl < 5`)
- `-5` packet is shorter than `ihl * 4`

Parsing requirements:
- Header fields are network byte order (big-endian).
- `version` and `ihl` come from the first byte.
- `flags` are the top 3 bits of bytes 6-7.
- `fragment_offset` is the low 13 bits of bytes 6-7.

This mirrors packet-inspection and protocol-debug tasks common in networking interviews.
