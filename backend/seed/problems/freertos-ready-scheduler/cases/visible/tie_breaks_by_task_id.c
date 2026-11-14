rtos_init();
task_create(6, 2);
task_create(4, 2);
task_create(5, 2);
case_passed = schedule_next() == 4;
