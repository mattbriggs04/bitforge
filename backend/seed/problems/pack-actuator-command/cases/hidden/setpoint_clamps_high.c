uint32_t word = pack_command(1, 2, true, 3000);
command_t out;
unpack_command(word, &out);
case_passed = out.setpoint == 2047;
