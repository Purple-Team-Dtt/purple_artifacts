#include <stdio.h>
#include <sys/prctl.h>
#include <string.h>
#include <unistd.h>

int main() {
    char name[] = "fake_kworker";

    prctl(PR_SET_NAME, name, 0, 0, 0);

    printf("Process name changed to: %s\n", name);

    sleep(60); // mantener el proceso vivo para verlo en ps/top
    return 0;
}
