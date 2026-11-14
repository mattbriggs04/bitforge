uint8_t frame[] = {0xA5u, 0x00u, 0x03u, 0x11u, 0x22u, 0x33u};
case_passed = frame_payload_len(frame, sizeof(frame)) == 3;
