package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
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

		coinData, coinErr = fetchData()
		if coinErr != nil {
			fmt.Printf("Error %v", coinErr)
		} else {
			return
		}
	}()

	// fetch data from Binance
	go func() {
		defer wg.Done()
		binanceData, binanceErr = fetchBinanceData()
		if binanceErr != nil {
			fmt.Printf("Error %v", binanceErr)
		} else {
			return
		}
	}()

	wg.Wait()

	if coinErr != nil {
		fmt.Printf("Error fetching data: %s", coinErr.Error())
	}
	if binanceErr != nil {
		fmt.Printf("Error fetching data: %s", binanceErr.Error())
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
		errMsg := fmt.Sprintf("Error creating HTTP request: %v", err)
		fmt.Println(errMsg)
		return CoinData{}, fmt.Errorf(errMsg)
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
		errMsg := fmt.Sprintf("Error sending request to CoinMarketCap %v", err)
		fmt.Println(errMsg)
		return CoinData{}, fmt.Errorf(errMsg)
	}

	respBody, _ := io.ReadAll(resp.Body)
	var coinData CoinData
	err = json.Unmarshal(respBody, &coinData)
	if err != nil {
		errMsg := fmt.Sprintf("Error unmarshaling JSON from CoinMarketCap: %v", err)
		fmt.Println(errMsg)
		return CoinData{}, fmt.Errorf(errMsg)
	}
	return coinData, nil
}

func fetchBinanceData() (BinanceData, error) {
	uri := "https://api.binance.com/api/v3/ticker/price?symbol=BTCUSDT"

	client := &http.Client{}

	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		errMsg := fmt.Sprintf("Error creating HTTP request: %v", err)
		fmt.Println(errMsg)
		return BinanceData{}, fmt.Errorf(errMsg)
	}
	resp, err := client.Do(req)
	if err != nil {
		errMsg := fmt.Sprintf("Error sending HTTP request to BinanceAPI: %v", err)
		fmt.Println(errMsg)
		return BinanceData{}, fmt.Errorf(errMsg)
	}

	respBody, _ := io.ReadAll(resp.Body)
	var coinData BinanceData
	err = json.Unmarshal(respBody, &coinData)
	if err != nil {
		errMsg := fmt.Sprintf("Error unmarshalling JSON from BinanceAPI: %v", err)
		fmt.Println(errMsg)
		return BinanceData{}, fmt.Errorf(errMsg)
	}
	return coinData, nil
}

func writeDataToCSV(data CoinData, binData BinanceData, filename string) error {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Can't open the file", err)
		return err
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
	timeNow := t.Format("2006-01-02 3:4:5 pm")

	for _, btc := range data.Data[0].Quote {
		price := fmt.Sprintf("%.2f", btc.Price)
		writer.Write([]string{timeNow, price, binData.Price})
		fmt.Println("Date:", timeNow, "BTC Price: $", price)
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
