This package is a fork of https://github.com/bouk/monkey.

As you may expect, this package is unsafe and fragile and probably
crash you program, it is only recommended for testing usage.

Notes

1. Monkey sometimes fails to patch a function if inlining is enabled.
   Try running your tests with inlining disabled, for example:
   `-gcflags=-l` (below go1.10) or `-gcflags=all=-l` (go1.10 and above).
   The same command line argument can also be used for build.
2. Monkey won't work on some security-oriented operating system that
   don't allow memory pages to be both write and execute at the same time.
   With the current approach there's not really a reliable fix for this.
3. Monkey is super unsafe, be sure you know what you are doing.
