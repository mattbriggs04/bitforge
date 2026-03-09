const char input[] = {'a', 'b', '\0', 'z', 'z', '\0'};
size_t got = bf_strlen(input);
size_t expected = 2;
case_passed = (got == expected);
