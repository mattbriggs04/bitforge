uint8_t packet[20] = {
    0x44, 0x00, 0x00, 0x14,
    0, 0, 0, 0,
    64, 17, 0, 0,
    1, 1, 1, 1,
    2, 2, 2, 2
};
ipv4_header_t out;
int rc = parse_ipv4_header(packet, sizeof(packet), &out);
case_passed = (rc == -4);
