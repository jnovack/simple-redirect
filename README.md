# simple-redirect

**simple-redirect** is my answer to a large number of requests I get to forward
an old url to another website while we transition the application.

Most of the time, one cannot simply add a CNAME to the new url because in order
for TLS to negotiate a CNAME, that name MUST be on the destination application's
certificate.

To work around this issue, and begin deprecation of the old URL, a 301 or 302
status code must be presented to the client.
