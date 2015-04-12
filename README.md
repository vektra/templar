templar
=======

[![Build Status](https://travis-ci.org/vektra/templar.svg?branch=master)](https://travis-ci.org/vektra/templar)

HTTP APIs, they're everywhere. But they have a serious problem: their
sychronous nature means that code using them stalls while waiting
for a reply.

This means that your apps uptime and reliability are intertwined with
whatever HTTP APIs, especially SaaS ones, you use.

templar helps you control the problem.

It is a an HTTP proxy that provides advanced features to help you make
better use of and tame HTTP APIs.

## Installation

Directly via go: `go get github.com/vektra/templar/cmd/templar`

### Linux

* [i386](https://bintray.com/artifact/download/evanphx/templar/templar-linux-386.tar.gz)
* [amd64](https://bintray.com/artifact/download/evanphx/templar/templar-linux-amd64.tar.gz)

### Darwin

* [amd64](https://bintray.com/artifact/download/evanphx/templar/templar-darwin-amd64.tar.gz)

### Windows

* [i386](https://bintray.com/artifact/download/evanphx/templar/templar-windows-386.zip)
* [amd64](https://bintray.com/artifact/download/evanphx/templar/templar-windows-amd64.zip)

## Usage

templar functions like an HTTP proxy, allowing you use your favorite HTTP client
to easily send requests through it. Various languages have different HTTP clients
but many respect the `http_proxy` environment variable that you can set to the
address templar is running on.

Most HTTP clients in various programming languages have some configuration
to configure the proxy directly as well. Nearly all of them do, just check
the docs.

### HTTPS

Many HTTP APIs located in SaaS products are available only via HTTPS. This is a
good thing though it makes templar's job a little harder. We don't want to a client
to use CONNECT because then we can't provide any value. So to interact with these APIs,
use the `X-Templar-Upgrade` header. Configure your client to talk to the API
as normal http but include `X-Templar-Upgrade: https` and templar will be able
manage your requests and still talk to the https service!

### Examples

Do a request through templar, no timeout, no caching:

`curl -x http://localhost:9224 http://api.openweathermap.org/data/2.5/weather?q=Los+Angeles,CA`


Now add some caching in, caching the value for a minute at a time:

`curl -x http://localhost:9224 -H "X-Templar-Cache: eager" -H "X-Templar-CacheFor: 1m" 'http://api.openweathermap.org/data/2.5/weather?q=Los+Angeles,CA'`


## Features

### Timeouts

It's important that timeouts are used when accessing a synchronous API like an
HTTP endpoint. It's not uncommon for upstream APIs to have no
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

There are 4 caches available presently:

* Memory (the default)
* Memcache
* Redis
* Groupcache

The later 3 are used only if configure on the command line.

In the future, the plan is to name the caches and allow requests to say which
caching backend they'd like to use. Currently they all use the same one.

### Stats generation

Tracking what APIs are used and how well they're performing is critical to
understanding. When requests flow through templar, it can generate metrics
about those requests and send them to statsd.

Just specify a statsd host via `-statsd` and templar will start sending them.

We'll support more metrics backends in the future.

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

A number of headers take time durations, for instances, 30 seconds. These use the simple "(number)(unit)" parser, so for 1 second, use `1s` and 5 minutes use `5m`. Units supported are: `ns`, `us`, `ms`, `s`, `m`, and `h`.

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
* Request stream inspection
* Fire-and-forget requests
* Return response via AMQP
