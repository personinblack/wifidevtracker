# Option 1 Websockets

Kismet's websocket API is slow but can be much more performant than sending http requests
every fucking second. Even though the "rate": 1; it does not update every second, only
when it really wants to. This behaviour might be connected to
"kismet.device.base.last_time" not being up to date with the one we use in these examples.

```bash
websocat "ws://192.168.1.2:2501/devices/monitor.ws?KISMET=ABD87C35D5D4F9C02C03279381D673A0"
# ps: websocat does not resolve "localhost".

{"monitor": "A4:50:46:3B:4F:4D", "request": 31337, "rate": 1, "fields": ["kismet.device.base.packets.rrd/kismet.common.rrd.last_time"]}
# "request" id can be used to cancel the subscription.
```

# Option 2 Get Requests

HTTP requests can be send every second and there for can be much more up to date. But
at what cost? By monitoring system usage while sending requests every second; we can
identify if it can get the system clogged.

```bash
curl "localhost:2501/devices/by-mac/A4:50:46:3B:4F:4D/devices.json?KISMET=ABD87C35D5D4F9C02C03279381D673A0" | grep "kismet.common.rrd.last_time"
```
