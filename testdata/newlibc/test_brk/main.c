#include <stdlib.h>

int main(void)
{
	if (malloc(0) == NULL)
		return -1;
	return 0;
}
