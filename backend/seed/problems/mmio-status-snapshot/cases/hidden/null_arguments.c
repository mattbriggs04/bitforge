volatile uint32_t status = 0;
volatile uint32_t clear = 0;
status_snapshot_t out = {0};
int a = read_status(0, &clear, &out);
int b = read_status(&status, &clear, 0);
case_passed = a == -1 && b == -1;
