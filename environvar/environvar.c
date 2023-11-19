#include <assert.h>
#include <stdio.h>

// required to be declared by programs
extern char **environ;

int main()
{
    printf("demonstrates using environ directly ...\n");

    printf("&environ=%p &environ[0]=%p\n", &environ, &environ[0]);
    int count = 0;
    for (char **env_it = environ; *env_it != NULL; env_it++)
    {
        printf("environ[%d]=%p; first byte=%p: %s\n", count, env_it, *env_it, *env_it);
        count++;
    }
    printf("%d total variables\n", count);
}
