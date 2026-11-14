uint32_t word = 2u | (1u << 4) | (1u << 7) | (0xFF9u << 8);
command_t out;
unpack_command(word, &out);
case_passed = out.actuator_id == 2 &&
              out.mode == 1 &&
              out.enabled &&
              out.setpoint == -7;
