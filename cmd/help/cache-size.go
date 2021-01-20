package help

const CacheSizeUsage = `cache-size [--prefix <prefix>]`

const CacheSize = `
Display the disk space occupied by the cache. If the
"--prefix" flag is not provided, all cached ones will
be displayed. If provided, only the space occupied by
the cached data of the specified prefix will be displayed.

See alse: clean-cache`
