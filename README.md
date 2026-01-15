**URL Shortener **
A simple URL shortener service built in Go.It supports bulk URL shortening, redirects, and includes extensive unit tests to cover edge cases.

**Features**
-Shorten multiple URLs at once
-Redirect short URLs to original links
-In-memory storage
-Secure random short codes
-Fully tested with edge-case coverage

**Edge Cases Covered**
-Invalid HTTP methods
-Invalid or missing JSON
-Empty URL list
-Invalid / unsupported URLs
-Duplicate URLs
-Very long URLs
-Short code collisions
-Missing redirect IDs
