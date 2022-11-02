# wifidevtracker
Track wifi devices in your area and send notifications on certain circumstances

## Instructions
1. Install and configure [Kismet](https://www.kismetwireless.net/).
2. Run Kismet: `$ kismet --daemonize --no-logging --device-timeout=-1 -c wlp5s0`. <sub>`$ kismet --help` for more options.
3. Open Kismet web ui (default: [localhost:2501](http://localhost:2501)) and generate an *$API_KEY*.
4. Clone this repository and run `$ go install`.
5. Start tracking with `$ API_KEY=YOUR_KISMET_API_KEY wifidevtracker` command. <sub>`$GOPATH/bin` has to be in your $PATH.
6. Edit default configuration file `$ nano $XDG_CONFIG_HOME/wifidevtracker/config.json` and restart tracking.
7. Profit. Notifications should start coming.
