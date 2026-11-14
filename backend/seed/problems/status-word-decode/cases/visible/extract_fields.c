uint16_t word = (uint16_t)(1u | (1u << 1) | (9u << 4) | (18u << 8));
status_t out;
decode_status(word, &out);
case_passed = out.ready &&
              out.degraded &&
              !out.fault &&
              out.subsystem == 9 &&
              out.error_code == 18;
