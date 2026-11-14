status_t out;
decode_status(0, &out);
case_passed = !out.ready &&
              !out.degraded &&
              !out.fault &&
              out.subsystem == 0 &&
              out.error_code == 0 &&
              !status_faulted(0);
