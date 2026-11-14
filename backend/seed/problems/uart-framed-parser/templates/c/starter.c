#include <stdbool.h>
#include <stdint.h>

#define UART_SYNC 0x7Eu
#define UART_MAX_PAYLOAD 4

typedef struct {
    uint8_t len;
    uint8_t payload[UART_MAX_PAYLOAD];
} uart_frame_t;

typedef struct {
    uint8_t state;
    uint8_t len;
    uint8_t index;
    uint8_t checksum;
    uint8_t payload[UART_MAX_PAYLOAD];
} uart_parser_t;

void uart_init(uart_parser_t *parser) {
    parser->state = 0;
    parser->len = 0;
    parser->index = 0;
    parser->checksum = 0;
}

bool uart_feed(uart_parser_t *parser, uint8_t byte, uart_frame_t *out) {
    (void)parser;
    (void)byte;
    (void)out;
    return false;
}
