volatile uint32_t status = 2u | 4u | (7u << 4) | (0x11u << 8);
volatile uint32_t clear = 0;
status_snapshot_t out = {0};
int rc = read_status(&status, &clear, &out);
case_passed = rc == 0 &&
              out.fault &&
              out.overrun &&
              out.mode == 7 &&
              out.error_code == 0x11 &&
              clear == (2u | 4u);
