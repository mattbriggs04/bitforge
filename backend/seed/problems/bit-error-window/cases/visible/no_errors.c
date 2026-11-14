error_window_t result = analyze_errors(0x12345678u, 0x12345678u);
case_passed = result.total_errors == 0 &&
              result.window_start == 0 &&
              result.window_errors == 0;
