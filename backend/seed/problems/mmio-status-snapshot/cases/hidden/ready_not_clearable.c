volatile uint32_t status = 1u;
volatile uint32_t clear = 0xAAu;
status_snapshot_t out = {0};
int rc = read_status(&status, &clear, &out);
case_passed = rc == 0 && out.ready && clear == 0;
