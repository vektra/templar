templar
=======

HTTP APIs, they're everywhere. But they have a serious problem: their
sychronous nature means that code using them stalls while waiting
for a reply.

This means that your apps uptime and reliability and intertwinned with
whatever HTTP APIs, especially SaaS ones, you use.

templar helps you control the problem.

It is a an HTTP proxy that provides advanced features to help you make
better use of and tame HTTP APIs.

## Features

### Timeouts

It's important when accessing a synchronous API like an HTTP endpoint
that timeouts are used. It's not uncommon for upstream APIs to have no
timeouts to fulfill a request so that typically needs to be done on the client
side. Effect use of timeouts on these APIs will improve the robustness
of your own system.

For great discussion on this, check out Jeff Hodges thoughts on the topic:
* https://www.youtube.com/watch?v=BKqgGpAOv1w
* http://www.somethingsimilar.com/2013/01/14/notes-on-distributed-systems-for-young-bloods/

At present, templar does not enforce a default timeout, it needs to be set
per request via the `X-Templar-Timeout` header. The format is documented
below under Duration format.

### Request collapsing

Say that your app hits an HTTP endpoint at http://isitawesomeyet.com.
When you send those HTTP requests through templar, it will reduce the
number of requests to the external service to the bare minumum by combining
requests together. So if a request comes in while we're waiting on another
request to the same endpoint, we combine those requests together and
serve the same data to both. This improves response times and reduces
load on upstream servers.

### Caching

Templar can, if requested, cache upstream requests. By setting the
`X-Templar-Cache` header to either `fallback` or `eager`, templar
will cache responses to the endpoint and serve them back.

`fallback` will only use the cache if accessing the endpoint times out.
`eager` will use the cache if it exists first and always repopulate
from the endpoint when needed.

The `X-Templar-CacheFor` header time is used to control how long a cached
value will be used for. See Duration format below for how to specify the time.

There are 3 caches available presently:

* Memory (the default)
* Memcache
* Redis

The later 2 are used only if configure on the command line.

In the future, the plan is to name the caches and allow requests to say which
caching backend they'd like to use. Currently they all use the same one.

## Request categorization

Not all requests should use some of these features, for instance, request collapsing.
So templar includes a categorizer to identify requests that it should apply
additional handling to. It identifies a request as `stateless` or not. If
it is stateless, then things like request collapsing and caching can be used.

By default, only GET requests are treated as `stateless`. The `X-Templar-Category`
header allows the user to explicitly specify the category. The 2 valid values are
`stateful` and `stateless`.

Again, a stateless request is subject to the following additional handling:

* Request collapsing
* Caching

## Duration format

A number of headers take time durations, for instances, 30 seconds. These uses the simple "(number)(unit)" parser, so for 1 second, use `1s` and 5 minutes use `5m`. Units supported are: `ns`, `us`, `ms`, `s`, `m`, and `h`.

## Control Headers

Templar uses a number of headers to control how the requests are processed.

### X-Templar-Cache

Possible values:

* **eager**: Return a value from the cache before checking upstream
* **fallback**: Return a value from the cache only if the upstream has issues

### X-Templar-CacheFor

When caching, how long to cache the value for. If caching and this isn't set,
the default is used.

### X-Templar-Cached

Set on responses that are served from the cache.

### X-Templar-Category

Possible values:

* **stateless**: Process the request as stateless
* **stateful**: Process the request as stateful

### X-Templar-Timeout

Specifies how long to wait for the response before giving up.

### X-Templar-TimedOut

Set to `true` on a response when the request timed out.

### X-Templar-Upgrade

Possible values:

* **https**: When connecting to the upstream, switch to https


# Future features

* Automatic caching based on HTTP Expire headers
* Request throttling
* Multiple active caching backends
* Internal cluster caching via groupcache
* Request stream inspection
* Fire-and-forget requests
* Return response via AMQP
