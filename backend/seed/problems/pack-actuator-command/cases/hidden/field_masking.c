uint32_t word = pack_command(31, 15, false, 0);
command_t out;
unpack_command(word, &out);
case_passed = out.actuator_id == 15 && out.mode == 7;
