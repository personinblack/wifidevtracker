package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"
)

func IsZero(val interface{}) bool {
	valRef := reflect.ValueOf(val)
	return val == nil || !valRef.IsValid() || valRef.IsZero()
}

func OrElse(val, fallback interface{}) interface{} {
	if IsZero(val) {
		return fallback
	}

	return val
}

var RUNNING = true

var API_KEY = os.Getenv("API_KEY")
var URL = OrElse(os.Getenv("URL"), "http://localhost:2501")

type DeviceRecord struct {
	LastTime  int    `json:"kismet.device.base.last_time,omitempty"`
	FirstTime int    `json:"kismet.device.base.first_time,omitempty"`
	MacAddr   string `json:"kismet.device.base.macAddr,omitempty"`
	LastBssid string `json:"dot11.device.last_bssid,omitempty"`
	Tags      struct {
		Notes string `json:"notes,omitempty"`
	} `json:"kismet.device.base.tags,omitempty"`
	Channel string `json:"kismet.device.base.channel,omitempty"`
}

func FetchDevRec(macAddr string) (DeviceRecord, error) {
	resp, postErr := http.Get(
		fmt.Sprintf(
			"%s/devices/by-mac/%s/devices.json?KISMET=%s",
			URL,
			macAddr,
			API_KEY,
		),
	)

	if postErr != nil {
		return DeviceRecord{}, fmt.Errorf("Couldn't POST: %v", postErr)
	} else if resp == nil {
		return DeviceRecord{}, fmt.Errorf("Invalid response.")
	}

	decoder := json.NewDecoder(resp.Body)
	var devRecords []DeviceRecord
	decodeErr := decoder.Decode(&devRecords)
	if decodeErr != nil {
		return DeviceRecord{}, fmt.Errorf("There was an error: %s.", decodeErr)
	} else if len(devRecords) < 1 {
		return DeviceRecord{}, fmt.Errorf("No record found.")
	}

	return devRecords[0], nil
}

type TrackCallback func(prevDevRec, curDevRec DeviceRecord, err error) (c0ntinue bool)

func TrackDev(macAddr string, interval int, cb TrackCallback) {
	var prevDevRec DeviceRecord
	for {
		curDevRec, err := FetchDevRec(macAddr)

		if !cb(prevDevRec, curDevRec, err) {
			break
		}

		prevDevRec = curDevRec
		time.Sleep(time.Duration(interval) * time.Millisecond)
	}
}

type ConfDevice struct {
	Bssid     string `json:"bssid,omitempty"`
	Tracking  bool   `json:"tracking,omitempty"`
	GoneFor   int    `json:"gone_for,omitempty"`   // In seconds. 0 means disabled.
	BackAfter int    `json:"back_after,omitempty"` // ^^
}

type Config struct {
	Interval int          `json:"interval,omitempty"` // In milliseconds.
	Devices  []ConfDevice `json:"devices"`
}

func ConfigDir() string {
	dir, err := os.UserConfigDir()

	if err != nil {
		panic(err)
	}

	dir = filepath.Join(dir, "wifidevtracker")

	err = os.Mkdir(dir, os.ModePerm)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			stat, errStat := os.Stat(dir)
			if errStat != nil {
				panic(errStat)
			} else if !stat.IsDir() {
				panic(err)
			}
		} else {
			panic(err)
		}
	}

	return dir
}

func ReadConfig() (Config, error) {
	conFile := filepath.Join(ConfigDir(), "config.json")
	rFile, err := os.ReadFile(conFile)
	if err != nil {
		return Config{}, err
	}

	var config Config
	err = json.Unmarshal(rFile, &config)

	if err != nil {
		return config, err
	}

	return config, nil
}

func WriteDefConfig() {
	conFile := filepath.Join(ConfigDir(), "config.json")

	if _, err := os.Stat(conFile); err == nil {
		return
	}

	config, err := json.MarshalIndent(Config{
		Interval: 1000,
		Devices: []ConfDevice{
			{
				Bssid:     "A4:50:46:3B:4F:4D",
				Tracking:  true,
				GoneFor:   0,
				BackAfter: 0,
			},
		},
	}, "", "\t")

	if err != nil {
		panic(err)
	}

	err = os.WriteFile(conFile, config, 0644)

	if err != nil {
		panic(err)
	}
}

func main() {
	if IsZero(API_KEY) {
		panic("You need to provide the API_KEY env variable!")
	}

	clog := log.New(os.Stdout, "", log.Ltime|log.Ldate|log.Lshortfile|log.Lmsgprefix)

	WriteDefConfig()
	config, cErr := ReadConfig()

	if cErr != nil {
		panic(cErr)
	}

	for _, dev := range config.Devices {
		if !dev.Tracking {
			continue
		}

		go TrackDev(
			dev.Bssid,
			config.Interval,
			func(prevDevRec, curDevRec DeviceRecord, err error) bool {
				if err != nil {
					if !strings.Contains(err.Error(), "No record found.") {
						clog.Println(err)
					}
					return true
				}

				prettyRec, _ := json.MarshalIndent(curDevRec, "", "\t")
				if !IsZero(prevDevRec) {
					timeDiff := curDevRec.LastTime - prevDevRec.LastTime
					if timeDiff > 0 {
						if dev.GoneFor > 0 {
							x := timeDiff - dev.GoneFor

							if x >= 0 && x <= 2 {
								// TODO: Send Notification
							}
						}

						fmt.Println()
						clog.Printf("DevRecord: %v", string(prettyRec))
						clog.Printf("TimeDiff in secs: %v\n", timeDiff)
					} else {
						fmt.Print(".")
					}
				} else {
					fmt.Println()
					clog.Printf("DevRecord: %v", string(prettyRec))
				}
				return true
			},
		)
	}

	for RUNNING {
		time.Sleep(time.Second)
	}

	fmt.Println()
	clog.Println("Goodbye!")
}
