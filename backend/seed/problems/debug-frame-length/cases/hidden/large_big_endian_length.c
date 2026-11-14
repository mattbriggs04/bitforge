uint8_t frame[3 + 258];
frame[0] = 0xA5u;
frame[1] = 0x01u;
frame[2] = 0x02u;
for (size_t i = 3; i < sizeof(frame); i++) {
    frame[i] = (uint8_t)i;
}
case_passed = frame_payload_len(frame, sizeof(frame)) == 258;
