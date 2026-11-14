error_window_t result = analyze_errors(0u, 0xFF000000u);
case_passed = result.total_errors == 8 &&
              result.window_start == 24 &&
              result.window_errors == 8;
