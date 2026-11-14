#include <stddef.h>
#include <stdint.h>

int frame_payload_len(const uint8_t *frame, size_t len) {
    if (len < 2) {
        return -1;
    }
    if (frame[0] != 0xA5u) {
        return -2;
    }

    uint16_t payload_len = (uint16_t)(frame[1] | ((uint16_t)frame[2] << 8));
    if (payload_len > len) {
        return -3;
    }
    return (int)payload_len;
}
