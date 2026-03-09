unsigned char buf[] = {0x10, 0x20, 0x30};
void *ret = bf_memmove(buf, buf, sizeof(buf));
case_passed = (ret == buf) && buf[0] == 0x10 && buf[1] == 0x20 && buf[2] == 0x30;
