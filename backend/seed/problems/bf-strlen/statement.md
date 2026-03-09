Implement `size_t bf_strlen(const char *s)`.

Return the number of bytes in `s` before the first null terminator (`'\0'`).

This is a foundational embedded-C routine used in bootloaders, RTOS utilities, and memory-constrained firmware where understanding pointer traversal matters.

Notes:
- Count bytes, not Unicode code points.
- Stop at the first `\0` byte.
- Do not call the standard library `strlen`.
