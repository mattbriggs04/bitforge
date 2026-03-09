uint8_t packet[20] = {
    0x45, 0x00, 0x00, 0x14,
    0, 0, 0, 0,
    64, 17, 0, 0,
    1, 2, 3, 4,
    5, 6, 7, 8
};
ipv4_header_t out;
int rc1 = parse_ipv4_header(NULL, sizeof(packet), &out);
int rc2 = parse_ipv4_header(packet, sizeof(packet), NULL);
case_passed = (rc1 == -1) && (rc2 == -1);
