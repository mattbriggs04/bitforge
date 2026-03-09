Implement an interrupt-driven button debounce state machine.

A timer ISR polls a GPIO input and calls:

- `void debounce_timer_isr(debounce_button_t *db, uint8_t pin_level)`

The debounce policy is an 8-bit shift register:
- Shift `history` left by 1.
- Insert latest sampled bit (`pin_level ? 1 : 0`) into bit0.
- If `history == 0xFF` and stable state was low, raise a rising-edge event.
- If `history == 0x00` and stable state was high, raise a falling-edge event.

You must also implement:
- `void debounce_init(debounce_button_t *db, uint8_t initial_level)`
- `uint8_t debounce_take_rising_edge(debounce_button_t *db)`
- `uint8_t debounce_take_falling_edge(debounce_button_t *db)`

`debounce_take_*` should return `1` once when an edge event is pending and clear that event.

This mimics firmware interview questions around timer-driven GPIO sampling and noisy switch handling.
