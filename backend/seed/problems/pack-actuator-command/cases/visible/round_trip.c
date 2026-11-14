uint32_t word = pack_command(9, 6, false, -123);
command_t out;
unpack_command(word, &out);
case_passed = out.actuator_id == 9 &&
              out.mode == 6 &&
              !out.enabled &&
              out.setpoint == -123;
