// 1. make request to get the price
// 2. open a csv file to save the price that i got.
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

func main() {
	uri := "https://pro-api.coinmarketcap.com/v1/cryptocurrency/listings/latest"

	client := &http.Client{}
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		log.Fatalf("Error: %v", err)
		os.Exit(1)
	}

	q := url.Values{}
	q.Add("start", "1")
	q.Add("limit", "1")
	q.Add("convert", "USD")
	req.Header.Set("Accepts", "application/json")
	req.Header.Add("X-CMC_PRO_API_KEY", os.Getenv("COIN_MARKET_CAP_TOKEN"))
	req.URL.RawQuery = q.Encode()

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
		return
	}

	parsedTime, err := time.Parse(time.RFC3339Nano, coinData.Status.Timestamp)
	if err != nil {
		fmt.Println("Failed to parse timestamp from API response.")
		return
	}

	shortTime := parsedTime.Format("2006-01-02 15:04")
	f, err := os.OpenFile("btc_prices.csv", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Can't open the file", err)
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()
	columnName := []string{"BTC Price in USD"}
	fileInfo, err := f.Stat()
	if err != nil {
		fmt.Println("Can't open the file", err)
		return
	}
	isEmpty := fileInfo.Size() == 0

	if isEmpty {
		w.Write(columnName)
		if err != nil {
			fmt.Println("Error writing columns: ", err)
			return
		}
	}

	if len(coinData.Data) > 0 {
		for _, v := range coinData.Data[0].Quote {
			price := fmt.Sprintf("%.2f", v.Price)
			cprice := fmt.Sprint(price)
			p := []string{fmt.Sprint(price)}
			w.Write(p)
			fmt.Println("Date:", shortTime, "BTC Price: $", cprice)
		}
	}
}
