uart_parser_t parser;
uart_frame_t frame;
uart_init(&parser);
uint8_t stream[] = {0x7Eu, 0x00u, 0x7Eu, 0x02u, 0xABu, 0xCDu, 0x64u};
bool emitted = false;
for (size_t i = 0; i < sizeof(stream); i++) {
    emitted = uart_feed(&parser, stream[i], &frame);
}
case_passed = emitted && frame.len == 2 && frame.payload[0] == 0xAB && frame.payload[1] == 0xCD;
