counter_reset(3);
counter_dec();
case_passed = counter_get() == 0;
