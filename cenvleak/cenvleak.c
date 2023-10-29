#include <assert.h>
#include <fcntl.h>
#include <limits.h>
#include <malloc.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <sys/resource.h>
#include <unistd.h>

static const char STATM_PATH[] = "/proc/self/statm";
static const size_t MAX_STATM_SIZE = 4096;
// see sysconf(_SC_PAGESIZE) but this is close enough
static const size_t INCORRECT_PAGE_SIZE = 4096;

// Returns the process's RSS in bytes by reading /proc/self/statm. Returns -1 on error. This assumes
// 4 kiB page sizes, which is not strictly true.
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

#ifdef __GLIBC__
// glibc provides mallinfo2()
struct malloc_stats
{
    struct mallinfo2 glibc_mallinfo2;
};

struct malloc_stats get_malloc_stats()
{
    struct malloc_stats result = {0};
    result.glibc_mallinfo2 = mallinfo2();
    return result;
}

void print_malloc_diff(struct malloc_stats before, struct malloc_stats after)
{
    printf("malloc info total allocated space bytes before=%zu after=%zu diff=%zu "
           "(mallinfo2.uordblks)\n",
           before.glibc_mallinfo2.uordblks, after.glibc_mallinfo2.uordblks,
           after.glibc_mallinfo2.uordblks - before.glibc_mallinfo2.uordblks);
}

#else
// musl does not provide malloc stats
struct malloc_stats
{
};

struct malloc_stats get_malloc_stats()
{
    struct malloc_stats result = {};
    return result;
}

void print_malloc_diff(struct malloc_stats before __attribute__((unused)),
                       struct malloc_stats after __attribute__((unused)))
{
    printf("malloc info not supported\n");
}
#endif

int main()
{
    printf("demonstrates that glibc setenv() never frees memory (musl does). Calling "
           "setenv/unsetenv ...\n");

    // This is not strictly true, but I don't have access to systems where it is not
    assert(sysconf(_SC_PAGESIZE) == INCORRECT_PAGE_SIZE);

    static const int NUM_TO_ALLOCATE = 10000;

    int before_rss_bytes = rss_bytes();
    assert(before_rss_bytes >= 0);

    struct malloc_stats before_malloc_stats = get_malloc_stats();

    // length must be at least 9 + length of NUM_TO_ALLOCATE; make it huge to be safe
    char env_name_buf[256];
    for (int i = 0; i < NUM_TO_ALLOCATE; i++)
    {
        int bytes_written = snprintf(env_name_buf, sizeof(env_name_buf), "env_var_%d", i);
        assert(0 < bytes_written && bytes_written < (int)sizeof(env_name_buf));
        int result = setenv(env_name_buf, "env_value", 1);
        assert(result == 0);
        result = unsetenv(env_name_buf);
        assert(result == 0);
    }

    int after_rss_bytes = rss_bytes();
    assert(after_rss_bytes >= 0);

    struct malloc_stats after_malloc_stats = get_malloc_stats();

    printf("RSS bytes before=%d after=%d diff=%d\n", before_rss_bytes, after_rss_bytes,
           after_rss_bytes - before_rss_bytes);
    print_malloc_diff(before_malloc_stats, after_malloc_stats);
}
