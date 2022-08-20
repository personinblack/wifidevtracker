package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"reflect"
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
var FIELD_SIMP, _ = json.Marshal(map[string][][]string{"fields": {
	{"kismet.device.base.last_time", "last_seen"},
}})

type DeviceRecord struct {
	LastSeen int `json:"last_seen,omitempty"`
}

func FetchDevRec(macAddr string) (DeviceRecord, error) {
	resp, postErr := http.Post(
		fmt.Sprintf(
			"%s/devices/by-mac/%s/devices.json?KISMET=%s",
			URL,
			macAddr,
			API_KEY,
		), "application/json",
		bytes.NewBuffer(FIELD_SIMP),
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

type callback func(prevDevRec, curDevRec DeviceRecord, err error) (c0ntinue bool)

func TrackDev(macAddr string, cb callback) {
	var prevDevRec DeviceRecord
	for {
		curDevRec, err := FetchDevRec(macAddr)

		if !cb(prevDevRec, curDevRec, err) {
			break
		}

		prevDevRec = curDevRec
		time.Sleep(time.Second)
	}
}

func main() {
	if API_KEY == "" {
		panic("You need to provide the API_KEY env variable!")
	}

	go TrackDev(
		"A4:50:46:3B:4F:4D",
		func(prevDevRec, curDevRec DeviceRecord, err error) bool {
			if err != nil {
				log.Println(err)
				RUNNING = false
				return false
			}

			if IsZero(prevDevRec) {
				timeDiff := curDevRec.LastSeen - prevDevRec.LastSeen
				if timeDiff > 0 {
					fmt.Printf("LastSeen: %v\n", curDevRec.LastSeen)
					fmt.Printf("TimeDiff in secs: %v\n", timeDiff)
				}
			} else {
				fmt.Printf("LastSeen: %v\n", curDevRec.LastSeen)
			}
			return true
		},
	)

	for RUNNING {
	}
}
