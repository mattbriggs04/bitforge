counter_reset(4);
counter_inc();
counter_inc();
counter_dec();
counter_inc();
case_passed = counter_get() == 2;
