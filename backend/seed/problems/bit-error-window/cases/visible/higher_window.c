uint32_t mask = 0xFFu << 12;
error_window_t result = analyze_errors(0u, mask);
case_passed = result.total_errors == 8 &&
              result.window_start == 12 &&
              result.window_errors == 8;
