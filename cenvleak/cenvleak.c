#include <stdio.h>
#include <sys/resource.h>
#include <assert.h>
#include <stdlib.h>
#include <stdint.h>
#include <fcntl.h>
#include <unistd.h>
#include <malloc.h>
#include <limits.h>

static const char STATM_PATH[] = "/proc/self/statm";
static const size_t MAX_STATM_SIZE = 4096;
// see sysconf(_SC_PAGESIZE) but this is close enough
static const size_t INCORRECT_PAGE_SIZE = 4096;

// Returns the process's RSS in bytes by reading /proc/self/statm. Returns -1 on error.
// This assumes 4 kiB page sizes, which is not strictly true.
int rss_bytes()
{
    int fd = open(STATM_PATH, O_RDONLY);
    if (fd < 0)
    {
        return fd;
    }

    // read the file and close it
    char buffer[MAX_STATM_SIZE];
    int read_result = read(fd, buffer, sizeof(buffer));
    int close_result = close(fd);
    if (read_result < 0)
    {
        return -1;
    }
    size_t statm_len = (size_t)read_result;
    assert(close_result == 0);

    // null terminate the string
    assert(statm_len < sizeof(buffer) - 1);
    buffer[statm_len] = 0;

    // find first space
    size_t space_index = 0;
    for (; space_index < statm_len; space_index++)
    {
        if (buffer[space_index] == ' ')
        {
            break;
        }
    }
    space_index += 1;
    assert(space_index < statm_len);

    int64_t rss_pages = strtol(&buffer[space_index], NULL, 10);
    assert(rss_pages != LONG_MIN && rss_pages != LONG_MAX);

    return rss_pages * INCORRECT_PAGE_SIZE;
}
int main()
{
    printf("demonstrates that setenv() never frees memory. Calling setenv/unsetenv ...\n");

    // This is not necessarily true, but I don't have access to systems where it is not
    assert(sysconf(_SC_PAGESIZE) == 4096);

    static const int NUM_TO_ALLOCATE = 10000;

    int before_rss_bytes = rss_bytes();
    assert(before_rss_bytes >= 0);

    struct mallinfo2 before_mallinfo = mallinfo2();

    char env_name_buf[4096];
    for (int i = 0; i < NUM_TO_ALLOCATE; i++)
    {
        int bytes_written = snprintf(env_name_buf, sizeof(env_name_buf), "env_var_%d", i);
        assert(bytes_written < (int)sizeof(env_name_buf));
        int result = setenv(env_name_buf, "env_value", 1);
        assert(result == 0);
        result = unsetenv(env_name_buf);
        assert(result == 0);
    }

    int after_rss_bytes = rss_bytes();
    assert(after_rss_bytes >= 0);

    struct mallinfo2 after_mallinfo = mallinfo2();

    printf("RSS bytes before=%d after=%d diff=%d\n",
           before_rss_bytes, after_rss_bytes, after_rss_bytes - before_rss_bytes);
    printf("malloc info total allocated space bytes before=%zu after=%zu diff=%zu\n",
           before_mallinfo.uordblks, after_mallinfo.uordblks,
           after_mallinfo.uordblks - before_mallinfo.uordblks);
}
