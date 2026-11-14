volatile uint32_t status = 0;
volatile uint32_t clear = 0xFFFFFFFFu;
status_snapshot_t out = {0};
int rc = read_status(&status, &clear, &out);
case_passed = rc == 0 &&
              !out.ready && !out.fault && !out.overrun &&
              out.mode == 0 && out.error_code == 0 &&
              clear == 0;
