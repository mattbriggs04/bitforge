uart_parser_t parser;
uart_frame_t frame;
uart_init(&parser);
uint8_t stream[] = {0x7Eu, 0x01u, 0x11u, 0x10u, 0x7Eu, 0x02u, 0x22u, 0x33u, 0x13u};
int emitted = 0;
uint8_t last_len = 0;
uint8_t last0 = 0;
uint8_t last1 = 0;
for (size_t i = 0; i < sizeof(stream); i++) {
    if (uart_feed(&parser, stream[i], &frame)) {
        emitted++;
        last_len = frame.len;
        last0 = frame.payload[0];
        last1 = frame.payload[1];
    }
}
case_passed = emitted == 2 && last_len == 2 && last0 == 0x22 && last1 == 0x33;
