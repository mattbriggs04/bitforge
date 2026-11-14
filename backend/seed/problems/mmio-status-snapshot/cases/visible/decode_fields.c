volatile uint32_t status = 1u | 2u | (5u << 4) | (0x2Au << 8);
volatile uint32_t clear = 0;
status_snapshot_t out = {0};
int rc = read_status(&status, &clear, &out);
case_passed = rc == 0 &&
              out.ready &&
              out.fault &&
              !out.overrun &&
              out.mode == 5 &&
              out.error_code == 0x2A &&
              clear == 2u;
