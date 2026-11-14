uint32_t mask = (0x0Fu << 4) | (0x0Fu << 20);
error_window_t result = analyze_errors(0u, mask);
case_passed = result.total_errors == 8 &&
              result.window_start == 0 &&
              result.window_errors == 4;
