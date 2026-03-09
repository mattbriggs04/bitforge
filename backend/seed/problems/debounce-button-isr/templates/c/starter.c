#include <stdint.h>

typedef struct {
    uint8_t history;
    uint8_t stable_state;
    uint8_t rose;
    uint8_t fell;
} debounce_button_t;

void debounce_init(debounce_button_t *db, uint8_t initial_level) {
    uint8_t level = initial_level ? 1u : 0u;
    db->history = level ? 0xFFu : 0x00u;
    db->stable_state = level;
    db->rose = 0u;
    db->fell = 0u;
}

void debounce_timer_isr(debounce_button_t *db, uint8_t pin_level) {
    (void)db;
    (void)pin_level;
}

uint8_t debounce_take_rising_edge(debounce_button_t *db) {
    uint8_t value = db->rose;
    db->rose = 0u;
    return value;
}

uint8_t debounce_take_falling_edge(debounce_button_t *db) {
    uint8_t value = db->fell;
    db->fell = 0u;
    return value;
}
