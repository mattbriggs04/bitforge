rtos_init();
task_create(0, 1);
task_create(7, 7);
task_set_ready(7, false);
case_passed = schedule_next() == 0;
