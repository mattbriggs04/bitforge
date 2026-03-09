uint8_t packet[20] = {
    0x45, 0xAB, 0x00, 0x14,
    0, 1, 0, 0,
    32, 6, 0x12, 0x34,
    1, 2, 3, 4,
    5, 6, 7, 8
};
ipv4_header_t out;
int rc = parse_ipv4_header(packet, sizeof(packet), &out);
case_passed = (rc == 0) && out.dscp == 0x2A && out.ecn == 0x3;
