uint8_t packet[19] = {
    0x45, 0x00, 0x00, 0x14,
    0, 0, 0, 0,
    64, 17, 0, 0,
    1, 2, 3, 4,
    5, 6, 7
};
ipv4_header_t out;
int rc = parse_ipv4_header(packet, sizeof(packet), &out);
case_passed = (rc == -2);
