# Telegraf Auth Proxy

For quite a while, I've been searching for a solution to properly authenticate remote metrics. I couldn't find one, so I've created one.

## The problem

Suppose you don't fully trust the network you're using to collect the metrics (which is fine). 
You want the clients (in our case [Telegraf](https://www.influxdata.com/time-series-platform/telegraf/) agents) to authenticate themselves.

Fortunately, Telegraf's HTTP [input](https://github.com/influxdata/telegraf/blob/release-1.28/plugins/inputs/http_listener_v2/README.md) and [output](https://github.com/influxdata/telegraf/blob/release-1.28/plugins/outputs/http/README.md) plugin supports basic authentication, but that's all it does. 
It only verifies basic authentication against a single, predefined username+password combination, and nothing else.

I prefer to use different authentication credentials for (at least) every host that can send in metrics. 
I can set this up with Nginx (or any other reverse proxy), but then again, it only verifies the basic auth header.

If I'm collecting the exact same metrics from multiple hosts, I think it makes perfect sense to put them into the same measurement in the same bucket, and use a tag (such as `host`) to differentiate between the hosts.
This is a practical solution, and telegraf supports it well, but there is a small problem:

The tag used to differentiate between hosts are set in the body of the request (part of the metrics), and it is not authenticated in any way.
If one of the credentials are leaked from any host (for example a malicious actor is gaining access just to one of the clients).
It can be used easily to impersonate other hosts, because the tag can be set to whatever since it's not checked.

This proxy tries to solve just that. It sits between the edge of your untrusted network and your target (another Telegraf instance, or InfluxDB directly). 
And does both basic authentication, and checks the body of each incoming metric, to make sure, that the `host` field matches the username, used for authentication.
This way, a single host credential cannot be used to submit metrics on behalf of any host.

## Deploy

TBD

## Configuration

TBD

## Endpoints

TBD