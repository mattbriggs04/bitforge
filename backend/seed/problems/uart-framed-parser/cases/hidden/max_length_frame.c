uart_parser_t parser;
uart_frame_t frame;
uart_init(&parser);
uint8_t stream[] = {0x7Eu, 0x04u, 1u, 2u, 3u, 4u, 0x00u};
bool emitted = false;
for (size_t i = 0; i < sizeof(stream); i++) {
    emitted = uart_feed(&parser, stream[i], &frame);
}
case_passed = emitted &&
              frame.len == 4 &&
              frame.payload[0] == 1 &&
              frame.payload[1] == 2 &&
              frame.payload[2] == 3 &&
              frame.payload[3] == 4;
