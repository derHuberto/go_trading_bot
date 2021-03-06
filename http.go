package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

//Takes an array with symbols to get the Historical Klines (Only from Binance)
func getHistoricalKlines(Symbols []string, datasetSize int, klineInterval string) {

	for _, symbol := range Symbols {
		APIurl, err := url.Parse("https://api.binance.com/api/v3/klines")

		if err != nil {
			log.Fatal(err)
		}

		URLparams := url.Values{}

		timeNow := time.Now()
		URLparams.Add("symbol", symbol)
		URLparams.Add("interval", klineInterval)

		//for dynamic times
		min, err := time.ParseDuration(klineInterval)

		if err != nil {
			switch klineInterval {
			case "1d":
				hours, _ := time.ParseDuration("24h")
				timeBack := int(hours.Hours()) * -1 * datasetSize

				URLparams.Add("startTime", fmt.Sprintf("%d", timeNow.Add(time.Duration(timeBack)*time.Hour).UnixMilli()))
				URLparams.Add("endTime", fmt.Sprintf("%d", timeNow.UnixMilli()))
			case "3d":
				hours, _ := time.ParseDuration("72h")
				timeBack := int(hours.Hours()) * -1 * datasetSize

				URLparams.Add("startTime", fmt.Sprintf("%d", timeNow.Add(time.Duration(timeBack)*time.Hour).UnixMilli()))
				URLparams.Add("endTime", fmt.Sprintf("%d", timeNow.UnixMilli()))
			case "1w":
				hours, _ := time.ParseDuration("168h")
				timeBack := int(hours.Hours()) * -1 * datasetSize

				URLparams.Add("startTime", fmt.Sprintf("%d", timeNow.Add(time.Duration(timeBack)*time.Hour).UnixMilli()))
				URLparams.Add("endTime", fmt.Sprintf("%d", timeNow.UnixMilli()))
			case "1M":
				hours, _ := time.ParseDuration("672h")

				timeBack := int(hours.Hours()) * -1 * datasetSize
				URLparams.Add("startTime", fmt.Sprintf("%d", timeNow.Add(time.Duration(timeBack)*time.Hour).UnixMilli()))
				URLparams.Add("endTime", fmt.Sprintf("%d", timeNow.UnixMilli()))
			default:
				log.Fatal("Not supported interval. All intervals are supported: 1m, 3m, 5m, 15m, 30m, 1h, 2h, 4h, 6h, 8h, 12h, 1d, 3d, 1w, 1M")
			}
		} else {
			timeBack := int(min.Minutes()) * -1 * datasetSize
			URLparams.Add("startTime", fmt.Sprintf("%d", timeNow.Add(time.Duration(timeBack)*time.Minute).UnixMilli()))
			URLparams.Add("endTime", fmt.Sprintf("%d", timeNow.UnixMilli()))
		}

		APIurl.RawQuery = URLparams.Encode()

		res, err := http.Get(APIurl.String())

		if err != nil {
			log.Fatal(err)
		}

		defer res.Body.Close()

		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			log.Fatal(err)
		}

		var ArrayRecv [][]interface{}

		jerr := json.Unmarshal(body, &ArrayRecv)
		if jerr != nil {
			log.Fatal(jerr)
		}

		for dfIndex, s := range symbols {
			if s == symbol {
				for _, oldKlines := range ArrayRecv {
					//convert json string to float for calc
					refloat, _ := strconv.ParseFloat(oldKlines[4].(string), 64)
					df[dfIndex].set = append(df[dfIndex].set, refloat)
				}
			}
		}

	}

}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func wsEndpoint(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}

	jsonMSG := status{Interval: Interval}

	for {

		if len(jsonMSG.Bot) == 0 {
			for _, elem := range BotArray {
				jsonMSG.Bot = append(jsonMSG.Bot, *elem)
				jsonMSG.OpenTrades = OpenTrades
			}
		} else {
			for index, elem := range BotArray {
				jsonMSG.Bot[index] = *elem
				jsonMSG.OpenTrades = OpenTrades
			}
		}

		jerr := ws.WriteJSON(jsonMSG)
		if jerr != nil {
			break
		}

		time.Sleep(20 * time.Second)
	}

}

func setupRoutes() {
	http.HandleFunc("/ws", wsEndpoint)
}

func Webserver() {
	setupRoutes()
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})
	http.ListenAndServe(":8080", nil)
}
