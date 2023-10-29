package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"
)

type CoinData struct {
	Data []struct {
		Quote map[string]struct {
			Price float64 `json:"price"`
		} `json:"quote"`
	} `json:"data"`
}

type BinanceData struct {
	Price string `json:"price"`
}

func fetchAndWriteDataToCSV() {
	var wg sync.WaitGroup
	var coinData CoinData
	var binanceData BinanceData
	var coinErr, binanceErr error

	wg.Add(2)

	// fetch data from CoinMarketCap
	go func() {
		defer wg.Done()

		data, err := fetchData()
		if err != nil {
			coinErr = err
		} else {
			coinData = data
		}
	}()

	// fetch data from Binance
	go func() {
		defer wg.Done()
		data, err := fetchBinanceData()
		if err != nil {
			binanceErr = err
		} else {
			binanceData = data
		}
	}()

	wg.Wait()

	if coinErr != nil {
		fmt.Println("Error fetching data:", coinErr)
	}
	if binanceErr != nil {
		fmt.Println("Error fetching data:", binanceErr)
	}
	// write the data to a CSV file
	err := writeDataToCSV(coinData, binanceData, "btc_prices.csv")
	if err != nil {
		fmt.Println("Error writing data to CSV:", err)
	}
}

func fetchData() (CoinData, error) {
	uri := "https://pro-api.coinmarketcap.com/v1/cryptocurrency/listings/latest"

	client := &http.Client{}

	// initiate http GET request againest the url
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		log.Fatalf("Error: %v", err)
		os.Exit(1)
	}

	// set Query Parameters and HTTP Headers
	q := url.Values{}
	q.Add("start", "1")
	q.Add("limit", "1")
	q.Add("convert", "USD")
	req.Header.Set("Accepts", "application/json")
	req.Header.Add("X-CMC_PRO_API_KEY", os.Getenv("COIN_MARKET_CAP_TOKEN"))
	req.URL.RawQuery = q.Encode()

	// sending the HTTP request
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending the request to CoinMarketCap!")
		os.Exit(1)
	}

	respBody, _ := io.ReadAll(resp.Body)
	var coinData CoinData
	err = json.Unmarshal(respBody, &coinData)
	if err != nil {
		fmt.Println("Error unmarshaling JSON:", err)
	}
	return coinData, nil
}

func fetchBinanceData() (BinanceData, error) {
	uri := "https://api.binance.com/api/v3/ticker/price?symbol=BTCUSDT"

	client := &http.Client{}

	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		fmt.Printf("Error: %v", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending the request to Binance API", err)
	}

	respBody, _ := io.ReadAll(resp.Body)
	var coinData BinanceData
	err = json.Unmarshal(respBody, &coinData)
	if err != nil {
		fmt.Println("Error unmarshalling JSON from Binance API: ", err)
	}
	return coinData, nil
}

func writeDataToCSV(data CoinData, binData BinanceData, filename string) error {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Can't open the file", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Adding CSV header
	columnName := []string{"Date", "CoinMarketCap Price", "Binance Price"}
	fileInfo, err := file.Stat()
	if err != nil {
		fmt.Println("Can't open the file", err)
	}
	// This will run if the file is not exist or is empty
	isEmpty := fileInfo.Size() == 0
	if isEmpty {
		writer.Write(columnName)
		if err != nil {
			fmt.Println("Error writing columns: ", err)
		}
	}

	// store local time to use it as output of the script and in the CSV
	t := time.Now()
	timenow := t.Format("2006-01-02 3:4:5 pm")

	for _, btc := range data.Data[0].Quote {
		price := fmt.Sprintf("%.2f", btc.Price)
		writer.Write([]string{timenow, fmt.Sprintf("%.2f", btc.Price), binData.Price})
		fmt.Println("Date:", timenow, "BTC Price: $", price)
	}

	fmt.Println("price of binance:", binData.Price)

	return nil
}

func main() {
	// Create a ticker that trigger every 5 minutes
	ticker := time.NewTicker(5 * time.Minute)

	// invoke the initial function
	go fetchAndWriteDataToCSV()
	//
	// Run script every 5 minutes
	for range ticker.C {
		go fetchAndWriteDataToCSV()
	}
}
