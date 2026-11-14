rtos_init();
int ok = task_create(1, 1) == 0 &&
         task_create(2, 3) == 0 &&
         task_create(3, 2) == 0;
case_passed = ok && schedule_next() == 2;
