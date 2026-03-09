char input[1024];
for (size_t i = 0; i < 1023; i++) {
    input[i] = 'x';
}
input[1023] = '\0';
size_t got = bf_strlen(input);
size_t expected = 1023;
case_passed = (got == expected);
