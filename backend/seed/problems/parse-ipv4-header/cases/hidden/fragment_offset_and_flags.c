uint8_t packet[20] = {
    0x45, 0x00, 0x00, 0x28,
    0xAA, 0x55, 0xBF, 0xFF,
    0x20, 0x01, 0x00, 0x00,
    172, 16, 0, 9,
    172, 16, 0, 10
};
ipv4_header_t out;
int rc = parse_ipv4_header(packet, sizeof(packet), &out);
case_passed = (rc == 0) && out.flags == 5 && out.fragment_offset == 8191;
