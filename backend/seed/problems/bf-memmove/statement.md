Implement `void *bf_memmove(void *dest, const void *src, size_t n)`.

`bf_memmove` copies bytes like `memcpy`, but must work correctly when source and destination regions overlap.

This pattern appears in packet reassembly, ring-buffer compaction, and in-place buffer editing.

Requirements:
- Copy exactly `n` bytes from `src` to `dest`.
- Correctly handle overlapping regions.
- Return the original `dest` pointer.
- Do not call libc `memmove`.
