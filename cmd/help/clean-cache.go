package help

const CleanCacheUsage = `clean-cache [--prefix <prefix>]`

const CleanCache = `
Recursively delete locally cached data. If the "--prefix"
flag is used, only the cached data beginning with the given
prefix will be deleted.

Cached data is generally used to speed up other commands,
and deleting them will not cause crashing effects.

If the data is inconsistent due to the cache, or the cache
takes up too much disk space (you can use "cache-size" to view
the space occupied by the cache), you can consider using
this command.

Another possibility is that because go-gendb's cached data is
all stored in plaintext, it may contain some sensitive information.
If you are worried that this will affect your security (such as
the risk of information being stolen by hackers) , You can use this
command to delete the cache, and use flags in other commands to
prohibit the use of the cache (usually it is disabled by default)

See alse: cache-size`
