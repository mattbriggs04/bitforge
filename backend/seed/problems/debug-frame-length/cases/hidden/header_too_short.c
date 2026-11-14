uint8_t frame[] = {0xA5u, 0x00u};
case_passed = frame_payload_len(frame, sizeof(frame)) == -1;
