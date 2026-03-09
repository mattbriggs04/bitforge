uint8_t packet[24] = {
    0x46, 0x03, 0x00, 0x30,
    0x12, 0x34, 0x20, 0x01,
    0x3F, 0x11, 0xAB, 0xCD,
    10, 0, 0, 1,
    10, 0, 0, 2,
    0xDE, 0xAD, 0xBE, 0xEF
};
ipv4_header_t out;
int rc = parse_ipv4_header(packet, sizeof(packet), &out);
case_passed = (rc == 0) &&
    out.version == 4 &&
    out.ihl == 6 &&
    out.dscp == 0 &&
    out.ecn == 3 &&
    out.total_length == 0x0030 &&
    out.identification == 0x1234 &&
    out.flags == 1 &&
    out.fragment_offset == 1 &&
    out.ttl == 63 &&
    out.protocol == 17 &&
    out.header_checksum == 0xABCD &&
    out.src_addr == 0x0A000001 &&
    out.dst_addr == 0x0A000002;
