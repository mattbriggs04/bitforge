uint16_t word = 0xFFFFu;
status_t out;
decode_status(word, &out);
case_passed = out.ready &&
              out.degraded &&
              out.fault &&
              out.subsystem == 15 &&
              out.error_code == 63;
