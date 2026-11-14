uart_parser_t parser;
uart_frame_t frame;
uart_init(&parser);
uint8_t stream[] = {0x7Eu, 0x01u, 0xAAu, 0x00u, 0x7Eu, 0x01u, 0x55u, 0x54u};
int emitted = 0;
for (size_t i = 0; i < sizeof(stream); i++) {
    if (uart_feed(&parser, stream[i], &frame)) {
        emitted++;
    }
}
case_passed = emitted == 1 && frame.len == 1 && frame.payload[0] == 0x55;
