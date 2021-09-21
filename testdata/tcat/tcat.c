#include <unistd.h>

void main(void) {
    write(0, "Hello World\n", 12);
}
