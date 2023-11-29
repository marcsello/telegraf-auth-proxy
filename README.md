# Telegraf Tag Auth Proxy

[![Build Status](https://drone.k8s.marcsello.com/api/badges/marcsello/telegraf-tag-auth-proxy/status.svg)](https://drone.k8s.marcsello.com/marcsello/telegraf-tag-auth-proxy)

For quite a while, I've been searching for a solution to properly authenticate remote metrics. I couldn't find one, so
I've created one.

## The problem

Suppose you don't fully trust the network you're using to collect the metrics (which is fine).
You want the clients (in our case [Telegraf](https://www.influxdata.com/time-series-platform/telegraf/) agents) to
authenticate themselves.

Fortunately, Telegraf's
HTTP [input](https://github.com/influxdata/telegraf/blob/release-1.28/plugins/inputs/http_listener_v2/README.md)
and [output](https://github.com/influxdata/telegraf/blob/release-1.28/plugins/outputs/http/README.md) plugin supports
basic authentication, but that's all it does.
It only verifies basic authentication against a single, predefined username+password combination, and nothing else.

I prefer to use different authentication credentials for (at least) every host that can send in metrics.
I can set this up with Nginx (or any other reverse proxy), but then again, it only verifies the basic auth header.

If I'm collecting the exact same metrics from multiple hosts, I think it makes perfect sense to put them into the same
measurement in the same bucket, and use a tag (such as `host`) to differentiate between the hosts.
This is a practical solution, and telegraf supports it well, but there is a small problem:

The tag used to differentiate between hosts are set in the body of the request (part of the metrics), and it is not
authenticated in any way.
If one of the credentials are leaked from any host (for example a malicious actor is gaining access just to one of the
clients).
It can be used easily to impersonate other hosts, because the tag can be set to whatever since it's not checked.

This proxy tries to solve just that. It sits between the edge of your untrusted network and your target (another
Telegraf instance, or InfluxDB directly).
And does both basic authentication, and checks the body of each incoming metric, to make sure, that the `host` field
matches the username, used for authentication.
This way, a single host credential cannot be used to submit metrics on behalf of any host.

## Deploy

Get from DockerHub: `marcsello/telegraf-anti-cheat-proxy`

## Configuration

Aside from the htaccess file entries. This proxy can be configured trough envvars:

| name                     | default           | description                                                                | notes                                                     |
|--------------------------|-------------------|----------------------------------------------------------------------------|-----------------------------------------------------------|
| `DEBUG`                  | `false`           | Enable debug logs                                                          |                                                           |
| `BIND_ADDR`              | `:8000`           | Address to bind the http server                                            |                                                           |
| `PROXY_UPSTREAM_URL`     |                   | The upstream URL to forward the requests to if the authentication succeeds | This field is required                                    |
| `PROXY_UPSTREAM_TIMEOUT` | `0`               | Timeout for the request towards the upstream. Set 0 to disable.            |                                                           |
| `AUTH_TAG`               | `host`            | Tag to be checked to be the same as the username in basic auth.            |                                                           |
| `MAX_BODY_LEN`           | `10737418240`     | Maximum size for parsing the request body. Extra bytes are dropped.        | The default is about 10M                                  |
| `LOAD_PARSERS`           | `influx`          | Comma separated list of parsers to be loaded. At least one is needed.      | See [Data types and Endpoints](#data-types-and-endpoints) |
| `HTPASSWD_PATH`          | `.htpasswd`       | The path for the htpasswd file. Default is in the current working dir.     |                                                           |
| `BASIC_AUTH_REALM`       | `restricted-area` | The basic auth realm to be reported by the proxy by default.               |                                                           |


## Data types and Endpoints

Each datatype have it's own endpoint which can be invoked both via POST and PUT. The data type name is the endpoint
name.

For example: `/influx` consumes influx datatype.

All requests are forwarded to the same upstream address regardless their data type.

**Currently, only the influx data type is supported!** (others would require configuration which I was lazy to
implement)