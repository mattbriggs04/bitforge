#include <stddef.h>
#include <stdint.h>

typedef struct {
    uint8_t version;
    uint8_t ihl;
    uint8_t dscp;
    uint8_t ecn;
    uint16_t total_length;
    uint16_t identification;
    uint8_t flags;
    uint16_t fragment_offset;
    uint8_t ttl;
    uint8_t protocol;
    uint16_t header_checksum;
    uint32_t src_addr;
    uint32_t dst_addr;
} ipv4_header_t;

int parse_ipv4_header(const uint8_t *packet, size_t len, ipv4_header_t *out) {
    (void)packet;
    (void)len;
    (void)out;
    return -1;
}
