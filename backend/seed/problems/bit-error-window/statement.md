# Bit Error Window

Compare two 32-bit samples and report the worst 8-bit error window:

```c
error_window_t analyze_errors(uint32_t expected, uint32_t observed);
```

Definitions:

- A bit error is any bit where `expected` and `observed` differ.
- `total_errors` is the number of differing bits across all 32 bits.
- `window_start` is the lowest bit index of the 8-bit window with the most errors.
- `window_errors` is the number of errors in that window.
- If windows tie, choose the lowest `window_start`.

Valid window starts are `0..24`.
