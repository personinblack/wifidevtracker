Here is how we do it on command line:

```bash
curl -d 'json={"fields": [["kismet.device.base.last_time", "last_seen"]]}' "localhost:2501/devices/by-mac/A4:50:46:3B:4F:4D/devices.json?KISMET=ABD87C35D5D4F9C02C03279381D673A0"
```

This returns the last_time the device has been seen in Unix timestamp.
