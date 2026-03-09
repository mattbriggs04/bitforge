uint8_t packet[20] = {
    0x45, 0x00, 0x00, 0x3C,
    0x1C, 0x46, 0x40, 0x00,
    0x40, 0x06, 0xB1, 0xE6,
    0xC0, 0xA8, 0x01, 0x0A,
    0x08, 0x08, 0x08, 0x08
};
ipv4_header_t out;
int rc = parse_ipv4_header(packet, sizeof(packet), &out);
case_passed = (rc == 0) &&
    out.version == 4 &&
    out.ihl == 5 &&
    out.total_length == 60 &&
    out.identification == 0x1C46 &&
    out.flags == 2 &&
    out.fragment_offset == 0 &&
    out.ttl == 64 &&
    out.protocol == 6 &&
    out.header_checksum == 0xB1E6 &&
    out.src_addr == 0xC0A8010A &&
    out.dst_addr == 0x08080808;
