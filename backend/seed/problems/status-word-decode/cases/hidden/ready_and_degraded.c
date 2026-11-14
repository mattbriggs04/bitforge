uint16_t word = 0x0003u;
status_t out;
decode_status(word, &out);
case_passed = out.ready && out.degraded && !out.fault;
