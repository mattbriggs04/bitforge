uint8_t packet[22] = {
    0x46, 0x00, 0x00, 0x16,
    0, 0, 0, 0,
    64, 17, 0, 0,
    192, 168, 0, 1,
    192, 168, 0, 2,
    0xAA, 0xBB
};
ipv4_header_t out;
int rc = parse_ipv4_header(packet, sizeof(packet), &out);
case_passed = (rc == -5);
