uint8_t frame[] = {0x5Au, 0x00u, 0x01u, 0x99u};
case_passed = frame_payload_len(frame, sizeof(frame)) == -2;
