uart_parser_t parser;
uart_frame_t frame;
uart_init(&parser);
uint8_t stream[] = {0x00u, 0xFFu, 0x7Eu, 0x01u, 0x42u, 0x43u};
bool emitted = false;
for (size_t i = 0; i < sizeof(stream); i++) {
    emitted = uart_feed(&parser, stream[i], &frame);
}
case_passed = emitted && frame.len == 1 && frame.payload[0] == 0x42;
