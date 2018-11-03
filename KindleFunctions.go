package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/fogleman/gg"
)

// Kbd defines pressed or release key and it's state 0 release, 1 pressed
type Kbd struct {
	key   int
	state int
}

// KeyboardWorker waits for key and add it to channel
func KeyboardWorker() {
	var k Kbd
	var err0 error
	var err1 error
	for {
		cmd := exec.Command("/usr/bin/waitforkey")
		cmdOut, err := cmd.Output()
		if err != nil {
			log.Fatalln("waitforkey ended with ", err)
		}
		//fmt.Println("Keyboard :", string(cmdOut))
		// waitfor key returns string og two values key code (104, 105, ..) and state code (1 or 0) plus "\n"
		// convert it to Kbd structure and send to keyboard channel
		tmps := strings.Split(strings.TrimSuffix(string(cmdOut), "\n"), " ")
		k.key, err0 = strconv.Atoi(tmps[0])
		k.state, err1 = strconv.Atoi(tmps[1])
		// fmt.Println("errors", err, err0, err1)
		if err == nil && err0 == nil && err1 == nil {
			//fmt.Println("Added: Key:", k.key, " Pressed: ", k.state)
			keyboard <- k
		}
	}
}

// MenuWorker consumes key from channel and updates program states
// at this moment only exits
func MenuWorker(config Config) {
	// keyboard scan codes:
	// 158 back
	// 29 keyboard
	// 105 left
	// 106 right
	// 103 up
	// 108 down
	// 194 enter / middle
	// 139 menu
	// 102 Home
	// 109 right page back
	// 191 right page next
	// 193 left page back
	// 104 left page next

	for {
		// non blocking check if channel not empty
		select {
		case k, ok := <-keyboard:
			if ok { // not empty, then process the key
				// fmt.Println("Got Key:", k.key, " Pressed: ", k.state)
				switch k.key {
				case 158: // back pressed or released does not matter :)
					// fmt.Println("Pressed", k.key, " exiting")
					wg.Done()
				case 29: // Keyboard button
					if k.state == 1 { // pressed
						job(config)
					}
				default: // .... next function to be built in, next city, detailed forecast or whatsoever
					fmt.Println("Got Key:", k.key, " Pressed: ", k.state)
				}
			} else {
				// that should not happened, keyboard worker never ends
				log.Fatalln("Channel closed!")
			}
		default:
		}
	}
}

func clearDisplay() {
	cmd := exec.Command("eips", "-c")
	cmd.Run()

	cmd = exec.Command("eips", "-c")
	cmd.Run()
}

func showImage(imagePath string) {
	cmd := exec.Command("eips", "-f", "-g", imagePath)
	err := cmd.Run()
	if err != nil {
		log.Println(err)
	}
}

func renderErrorDisp(icon, message string) {
	const my = 75 // margin from top

	dc := gg.NewContext(600, 800)
	ClearPic(dc)

	// print ! icon
	if err := dc.LoadFontFace("./fonts/kindleweathersr.ttf", 200); err != nil {
		panic(err)
	}
	dc.DrawStringAnchored(icon, 300, my+200, 0.5, 0.5)

	if err := dc.LoadFontFace("./fonts/Robotosr.ttf", 40); err != nil {
		panic(err)
	}
	dc.DrawStringWrapped(message, 300, 450, 0.5, 0.5, 550, 1.5, gg.AlignCenter)
	dc.Stroke()

	SaveGrayPic(dc.Image(), picFile)
}

// CheckWiFi returns true if kindle wifi interface is up and connected
func CheckWiFi() bool {
	cmd := exec.Command("/usr/bin/lipc-get-prop", "com.lab126.wifid", "cmState")
	cmdOut, err := cmd.CombinedOutput()
	if err != nil {
		//panic(err)
		log.Fatalln("Error: ", err)
	}
	return strings.TrimSuffix(string(cmdOut), "\n") == "CONNECTED"
}

func reconnectWiFi() {
	// ifconfig wlan0 down/up
	cmd := exec.Command("/usr/bin/wpa_cli", "-i", "wlan0", "reconnect")
	err := cmd.Run()
	if err != nil {
		log.Println(err)
	}
	time.Sleep(time.Duration(2 * time.Second))
}

// InternetConnected checkes if working internet connection exist.
func InternetConnected() bool {
	_, err := http.Get("http://clients3.google.com/generate_204") // ("http://samples.openweathermap.org/")
	if err != nil {
		return false
	}
	return true
}

// CheckNetwork checks the wifi interface for being "CONNECT"
// tries to recover in case the network is in pending mode (after droping wifi from the router)
// and checks if there is an active internet connection
func CheckNetwork() error {
	var connected = false // activly change from error to no-error if there is no :)
	connected = CheckWiFi()
	if !connected {
		reconnectWiFi() // try to recover a wifi connection
	}
	if connected || CheckWiFi() { // check if either 1) connected or 2) recovered WiFi connection has an access to internet
		if !InternetConnected() {
			return errors.New("WiFi OK, but no Internet Connection")
		}
	} else {
		return errors.New("WiFi connection not available")
	}
	return nil // we wnet through the check path wifi and internet are both fine
}