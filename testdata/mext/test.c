void _write(int fd, const char *buf, int len) {
    register int a0 asm("a0") = fd;
    register const char *a1 asm("a1") = buf;
    register int a2 asm("a2") = len;
    register int a7 asm("a7") = 64;
    asm volatile("ecall" : : "r"(a0), "r"(a1), "r"(a2), "r"(a7));
}

void _exit(int code) {
    register int a0 asm("a0") = code;
    register int a7 asm("a7") = 93;
    asm volatile("ecall" : : "r"(a0), "r"(a7));
    __builtin_unreachable();
}

void print(const char *s) {
    int len = 0;
    while (s[len]) len++;
    _write(1, s, len);
}

void print_num(long n) {
    char buf[24];
    int i = 22;
    int neg = 0;

    if (n < 0) {
        neg = 1;
        n = -n;
    }

    buf[23] = '\0';
    do {
        buf[i--] = '0' + (n % 10);
        n /= 10;
    } while (n > 0);

    if (neg) buf[i--] = '-';

    print(&buf[i + 1]);
}

void _start() {
    long a = 7;
    long b = 6;
    long c, d, e;

    c = a * b;
    print("7 * 6 = ");
    print_num(c);
    print("\n");

    c = 100 / 7;
    print("100 / 7 = ");
    print_num(c);
    print("\n");

    c = 100 % 7;
    print("100 % 7 = ");
    print_num(c);
    print("\n");

    c = -50 / 7;
    print("-50 / 7 = ");
    print_num(c);
    print("\n");

    c = -50 % 7;
    print("-50 % 7 = ");
    print_num(c);
    print("\n");

    _exit(0);
}
