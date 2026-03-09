uint8_t packet[20] = {
    0x65, 0x00, 0x00, 0x14,
    0, 0, 0, 0,
    64, 17, 0, 0,
    1, 2, 3, 4,
    5, 6, 7, 8
};
ipv4_header_t out;
int rc = parse_ipv4_header(packet, sizeof(packet), &out);
case_passed = (rc == -3);
