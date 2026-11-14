counter_reset(3);
counter_inc();
counter_inc();
counter_inc();
counter_inc();
case_passed = counter_get() == 3;
