uint16_t word = (uint16_t)(63u << 8);
status_t out;
decode_status(word, &out);
case_passed = out.error_code == 63 &&
              !out.fault &&
              !status_faulted(word);
