const char input[] = {'\0', 'x', 'y', 'z'};
size_t got = bf_strlen(input);
size_t expected = 0;
case_passed = (got == expected);
