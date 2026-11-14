rtos_init();
int a = task_create(3, 1);
int b = task_create(3, 4);
int c = task_create(8, 1);
case_passed = a == 0 && b == -1 && c == -1 && schedule_next() == 3;
