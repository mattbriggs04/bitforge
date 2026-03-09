Implement `void *bf_memcpy(void *dest, const void *src, size_t n)`.

This routine copies bytes from `src` to `dest` and returns `dest`.

In firmware and boot code, this primitive is often implemented manually and must behave predictably for raw byte buffers.

Requirements:
- Copy exactly `n` bytes from `src` to `dest`.
- Return the original `dest` pointer.
- Do not call libc `memcpy`.

Note: Like standard `memcpy`, overlapping regions are undefined behavior and are not part of this problem.
