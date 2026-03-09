char input[300];
for (size_t i = 0; i < 299; i++) {
    input[i] = 'a';
}
input[255] = '\0';
size_t got = bf_strlen(input);
size_t expected = 255;
case_passed = (got == expected);
