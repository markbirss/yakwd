package main

import (
	"fmt"
	"time"

	owm "github.com/briandowns/openweathermap"
	"github.com/fogleman/gg"
)

// RenderWeatherDisp converts txt structure into a png picture
func RenderWeatherDisp(config *Config, displayTxt *displayTxtType) {
	const my = 75       // margin from top
	const dy = 25 + 200 // high of a column
	const ix = 70       // x of a midle of a weather icon in the firts column
	const dx = 200      // width of a column
	const wy = 70       // y of a midle of a weather icon in a first column
	const tx = 160      // x of a first temperature text
	const ty = 120      // y of a first temperature text
	const hy = -20      // move up header text
	const iy = 190      // y of the the detailed weather information
	const fy = 30       // y footer from the last line
	const fx = 20       // x from left and right for footer text

	// for i := 0; i < numDays; i++ {
	// 	for j := 0; j < timeZones; j++ {
	// 		fmt.Print(displayTxt[i].Temp[j], ", ")
	// 	}
	// 	fmt.Println("")
	// }

	dc := gg.NewContext(600, 800)
	//img := image.NewGray(image.Rect(0, 0, 600, 800))
	//dc := gg.NewContextForImage(img) // does not work, img will be anyway converted to full RGB and requires closing convertion into gray scale
	ClearPic(dc)

	// create a table 5x3 (header, forecast 3x3, footer)
	dc.SetLineWidth(1)
	for i := 0; i < 4; i++ {
		dc.DrawLine(0, my+float64(i*dy), 600, my+float64(i*dy))
	}
	dc.DrawLine(1*dx, 0, 1*dx, my+3*dy)
	dc.DrawLine(2*dx, 0, 2*dx, my+3*dy)

	for i := 0; i < numDays; i++ {
		// Print Temperature
		if err := dc.LoadFontFace(config.TxtFont, 65); err != nil {
			panic(err)
		}
		for j := 0; j < timeZones; j++ {
			if displayTxt.Icon[i][j] != "?" {
				dc.DrawStringAnchored(fmt.Sprintf("%-0.0f", displayTxt.Temp[i][j]), tx+float64(j)*dx, my+ty+float64(i)*dy, 1, 1)
			}
		}

		// add smaller °C at the end of each temp.
		if err := dc.LoadFontFace(config.TxtFont, 30); err != nil {
			panic(err)
		}
		for j := 0; j < timeZones; j++ {
			if displayTxt.Icon[i][j] != "?" {
				dc.DrawStringAnchored("°C", tx+float64(j)*dx, my+ty+float64(i)*dy, 0, 1)
			}
		}

		// add detail (small text) weather description
		if err := dc.LoadFontFace(config.TxtFont, 16); err != nil {
			panic(err)
		}
		for j := 0; j < timeZones; j++ {
			if displayTxt.Icon[i][j] != "?" {
				dc.DrawStringWrapped(displayTxt.Description[i][j], dx/2+float64(j)*dx, my+iy+float64(i)*dy, 0.5, 0.5, 180, 1.5, gg.AlignCenter)
			}
		}

		// print weather icon
		if err := dc.LoadFontFace(config.IconFont, 100); err != nil {
			panic(err)
		}

		for j := 0; j < timeZones; j++ {
			dc.DrawStringAnchored(displayTxt.Icon[i][j], ix+float64(j)*dy, my+wy+float64(i)*dy, 0.5, 0.5)
		}

	}
	// other static text
	if err := dc.LoadFontFace(config.TxtFont, 30); err != nil {
		panic(err)
	}
	dc.DrawStringAnchored("Morning", dx/2+0*dx, my+hy, 0.5, 0)
	dc.DrawStringAnchored("Afternoon", dx/2+1*dx, my+hy, 0.5, 0)
	dc.DrawStringAnchored("Evening", dx/2+2*dx, my+hy, 0.5, 0)

	if err := dc.LoadFontFace(config.TxtFont, 24); err != nil {
		panic(err)
	}

	const maxCityLength = 20
	var strtmp string
	if len(displayTxt.City) > maxCityLength {
		strtmp = displayTxt.City[:maxCityLength]
	} else {
		strtmp = displayTxt.City
	}

	dc.DrawStringAnchored(strtmp+" @ "+displayTxt.TimeStamp, 250, my+3*dy+fy, 0.5, 0)

	// print Battery icon
	if err := dc.LoadFontFace(config.IconFont, 40); err != nil {
		panic(err)
	}
	dc.DrawStringAnchored(displayTxt.Batt, 600-3*fx, my+3*dy+fy, 0, 0.5)

	dc.Stroke()
	// dc.SavePNG("tmp.png") // not a gray scale picture - we do not need it
	SaveGrayPic(dc.Image(), picFile)
}

// ProcessWeatherData converts json structure from open weather into txt structure
// Selecting data from given time zones during the day
func ProcessWeatherData(config *Config, displayTxt *displayTxtType, w *owm.Forecast5WeatherData) {

	var weatherFontMapping = map[int]string{
		// openweather Main weather code to icon conversion
		200: "s", 201: "s", 202: "s", 210: "p", 211: "s", 212: "s", 221: "s", 230: "s", 231: "s", 232: "s",
		300: "k", 301: "k", 302: "k", 310: "k", 311: "k", 312: "k", 313: "k", 314: "k", 321: "k",
		500: "c", 501: "c", 502: "c", 503: "b", 504: "b", 511: "i", 520: "c", 521: "c", 522: "c", 531: "c",
		600: "r", 601: "r", 602: "r", 611: "f", 612: "r", 615: "u", 616: "u", 620: "r", 621: "r", 622: "r",
		701: "t", 711: "o", 721: "h", 731: "d", 741: "h", 751: "d", 761: "d", 762: "d", 771: "v", 781: "w",
		800: "n", 801: "j", 802: "j", 803: "a", 804: "x"}

	currentTime := time.Now().UTC()
	localTime := currentTime
	location, err := time.LoadLocation("Local")
	if err == nil {
		localTime = localTime.In(location)
	}

	// fmt.Println("UTC: ", currentTime)
	// fmt.Println("Loc: ", localTime)
	currentTime = localTime

	displayTxt.City = w.City.Name
	displayTxt.TimeStamp = currentTime.Format("15:04:05 Mon _2 Jan 2006")

	// time zones ranges are defined based on 3 hours blocks as the data are comming from open weather maps
	// Starting from 00:00, 03:00, 06:00, 09:00, 12, 15, 18, 21, 24=00:00
	// Current Check Points: 6=morning, 12=afternoon, 18-evening
	// Check points are compare with the open weather map time in UTC format !!!
	d1TimeZone1 := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), 6, 0, 0, 0, time.UTC)
	d1TimeZone2 := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), 12, 0, 0, 0, time.UTC)
	d1TimeZone3 := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), 18, 0, 0, 0, time.UTC)
	d2TimeZone1 := d1TimeZone1.AddDate(0, 0, 1) // tomorrow
	d2TimeZone2 := d1TimeZone2.AddDate(0, 0, 1)
	d2TimeZone3 := d1TimeZone3.AddDate(0, 0, 1)
	d3TimeZone1 := d1TimeZone1.AddDate(0, 0, 2) // the day after tomorrow
	d3TimeZone2 := d1TimeZone2.AddDate(0, 0, 2)
	d3TimeZone3 := d1TimeZone3.AddDate(0, 0, 2)

	// Convert all to unix int format to compare it with dt time from open weather maps
	d1T1U := d1TimeZone1.Unix() // day 1 Time zone 1 format Unix
	d1T2U := d1TimeZone2.Unix() // day 2 Time zone 2 format Unix
	d1T3U := d1TimeZone3.Unix() // ...
	d2T1U := d2TimeZone1.Unix()
	d2T2U := d2TimeZone2.Unix()
	d2T3U := d2TimeZone3.Unix()
	d3T1U := d3TimeZone1.Unix()
	d3T2U := d3TimeZone2.Unix()
	d3T3U := d3TimeZone3.Unix()

	var key owm.Forecast5WeatherList
	for i := len(w.List) - 1; i >= 0; i-- {

		forecastDay := -1
		timeZone := -1
		key = w.List[i]

		switch tt := key.Dt; {
		case (tt >= d1T1U) && (tt < d1T2U):
			forecastDay = 0
			timeZone = 0
		case (tt >= d1T2U) && (tt < d1T3U):
			forecastDay = 0
			timeZone = 1
		case (tt >= d1T3U) && (tt < d2T1U):
			forecastDay = 0
			timeZone = 2
		case (tt >= d2T1U) && (tt < d2T2U):
			forecastDay = 1
			timeZone = 0
		case (tt >= d2T2U) && (tt < d2T3U):
			forecastDay = 1
			timeZone = 1
		case (tt >= d2T3U) && (tt < d3T1U):
			forecastDay = 1
			timeZone = 2
		case (tt >= d3T1U) && (tt < d3T2U):
			forecastDay = 2
			timeZone = 0
		case (tt >= d3T2U) && (tt < d3T3U):
			forecastDay = 2
			timeZone = 1
		case (tt >= d3T3U):
			forecastDay = 2
			timeZone = 2
		}
		if (forecastDay != -1) && (timeZone != -1) {
			displayTxt.Temp[forecastDay][timeZone] = key.Main.Temp
			displayTxt.Description[forecastDay][timeZone] = key.Weather[0].Description
			displayTxt.Icon[forecastDay][timeZone] = weatherFontMapping[key.Weather[0].ID]
			if displayTxt.Icon[forecastDay][timeZone] == "" {
				displayTxt.Icon[forecastDay][timeZone] = "?"
			}
			//customLogf("Day: %d, Time: %d  ", forecastDay, timeZone)
			//fmt.Println("dt:", time.Unix(key.Dt, 0).UTC(), " dt: ", key.Dt)
		}
	}
	RenderWeatherDisp(config, displayTxt)
}

func getForecast5(config *Config) (*owm.Forecast5WeatherData, error) {
	w, err := owm.NewForecast("5", "c", "en", config.APIKey)
	if err != nil {
		return nil, err
	}
	// w.DailyByName("Albuquerque", 40)
	// better use City ID to get unique responce from open weather maps
	// run below command from your linux terminal to find id of your city
	// wget -qO - http://bulk.openweathermap.org/sample/city.list.json.gz | zcat | grep -i -B1 -A4 Albuquerque
	// and update the config.json file with the city ID

	w.DailyByID(config.CityIDTable[config.CityIDx], 40)
	forecast := w.ForecastWeatherJson.(*owm.Forecast5WeatherData)
	return forecast, err
}
