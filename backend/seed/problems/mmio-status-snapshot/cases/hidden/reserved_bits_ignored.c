volatile uint32_t status = 0xFFFF0008u | (3u << 4) | (0x55u << 8);
volatile uint32_t clear = 0;
status_snapshot_t out = {0};
int rc = read_status(&status, &clear, &out);
case_passed = rc == 0 &&
              !out.ready && !out.fault && !out.overrun &&
              out.mode == 3 &&
              out.error_code == 0x55 &&
              clear == 0;
