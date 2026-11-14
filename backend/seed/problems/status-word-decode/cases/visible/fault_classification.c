uint16_t word = (uint16_t)(1u << 2);
status_t out;
decode_status(word, &out);
case_passed = out.fault && status_faulted(word);
