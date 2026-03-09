Simulated Register Model

Assume a periodic timer ISR reads a GPIO input register and passes one bit to your code.

Example pseudo-flow:

  uint8_t pin = (GPIO_IDR >> BUTTON_PIN) & 0x1;
  debounce_timer_isr(&button_db, pin);

Your implementation is evaluated purely from the sampled `pin_level` sequence.
No hardware-specific headers are required.
