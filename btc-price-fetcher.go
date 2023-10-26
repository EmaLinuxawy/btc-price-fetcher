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
	"time"
)

type CoinData struct {
	Status struct {
		Timestamp string `json:"timestamp"`
	}
	Data []struct {
		Name  string `json:"name"`
		Quote map[string]struct {
			Price float64 `json:"price"`
		} `json:"quote"`
	} `json:"data"`
}

func fetchAndWriteDataToCSV() {
	// fetch data
	data, err := fetchData()
	if err != nil {
		fmt.Println("Error fetching data:", err)
	}

	// write the data to a CSV file
	err = writeDataToCSV(data, "btc_prices.csv")
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

func writeDataToCSV(data CoinData, filename string) error {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Can't open the file", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Adding CSV header
	columnName := []string{"Date", "BTC Price in USD"}
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

	// convert the timestamp to human readable
	parsedTime, err := time.Parse(time.RFC3339Nano, data.Status.Timestamp)
	if err != nil {
		fmt.Println("Failed to parse timestamp from API response.")
	}

	shortTime := parsedTime.Format("2006-01-02 15:04")

	for _, btc := range data.Data[0].Quote {
		price := fmt.Sprintf("%.2f", btc.Price)
		writer.Write([]string{shortTime, fmt.Sprintf("%.2f", btc.Price)})
		fmt.Println("Date:", shortTime, "BTC Price: $", price)
	}

	return nil
}

func main() {
	// Create a ticker that trigger every 5 minutes
	ticker := time.NewTicker(5 * time.Minute)

	// invoke the initial function
	fetchAndWriteDataToCSV()

	// Run script every 5 minutes
	for range ticker.C {
		fetchAndWriteDataToCSV()
	}
}
