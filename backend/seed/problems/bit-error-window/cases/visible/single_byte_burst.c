error_window_t result = analyze_errors(0x00000000u, 0x000000FFu);
case_passed = result.total_errors == 8 &&
              result.window_start == 0 &&
              result.window_errors == 8;
