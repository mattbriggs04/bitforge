uart_parser_t parser;
uart_frame_t frame;
uart_init(&parser);
uint8_t stream[] = {0x7Eu, 0x03u, 0x10u, 0x20u, 0x30u, 0x03u};
bool emitted = false;
for (size_t i = 0; i < sizeof(stream); i++) {
    emitted = uart_feed(&parser, stream[i], &frame);
}
case_passed = emitted &&
              frame.len == 3 &&
              frame.payload[0] == 0x10 &&
              frame.payload[1] == 0x20 &&
              frame.payload[2] == 0x30;
