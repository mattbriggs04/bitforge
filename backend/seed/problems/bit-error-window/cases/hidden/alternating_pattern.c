error_window_t result = analyze_errors(0xAAAAAAAAu, 0x55555555u);
case_passed = result.total_errors == 32 &&
              result.window_start == 0 &&
              result.window_errors == 8;
