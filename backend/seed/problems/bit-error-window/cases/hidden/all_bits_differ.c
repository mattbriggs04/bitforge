error_window_t result = analyze_errors(0u, 0xFFFFFFFFu);
case_passed = result.total_errors == 32 &&
              result.window_start == 0 &&
              result.window_errors == 8;
