package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"syscall"
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

func fetchData(ctx context.Context) (CoinData, error) {
	uri := "https://pro-api.coinmarketcap.com/v1/cryptocurrency/listings/latest"

	client := &http.Client{}

	// initiate http GET request againest the url
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		fmt.Printf("Error creating HTTP request: %v", err)
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
		fmt.Printf("Error sending request to CoinMarketCap %v", err)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return CoinData{}, fmt.Errorf("error reading response body: %w", err)
	}
	defer resp.Body.Close()

	var coinData CoinData
	if err = json.Unmarshal(respBody, &coinData); err != nil {
		return CoinData{}, fmt.Errorf("Error unmarshaling JSON from CoinMarketCap: %w", err)
	}

	if err := ctx.Err(); err != nil {
		return CoinData{}, err
	}
	return coinData, nil
}

func fetchBinanceData(ctx context.Context) (BinanceData, error) {
	uri := "https://api.binance.com/api/v3/ticker/price?symbol=BTCUSDT"

	client := &http.Client{}

	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		fmt.Printf("Error creating HTTP request: %v", err)

	}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error sending HTTP request to BinanceAPI: %v", err)

	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return BinanceData{}, fmt.Errorf("error reading response body: %w", err)
	}
	var coinData BinanceData
	err = json.Unmarshal(respBody, &coinData)
	if err != nil {
		fmt.Printf("Error unmarshalling JSON from BinanceAPI: %v", err)

	}

	if err := ctx.Err(); err != nil {
		return BinanceData{}, err
	}
	return coinData, nil
}

func fetchAndWriteDataToCSV(ctx context.Context, cancel context.CancelFunc) {
	var wg sync.WaitGroup

	coinDataCh := make(chan CoinData, 1)
	binanceDataCh := make(chan BinanceData, 1)
	errCh := make(chan error, 2)

	wg.Add(2)

	go func() {
		defer wg.Done()

		coinData, err := fetchData(ctx)
		if err != nil {
			errCh <- fmt.Errorf("Error fetching data from CoinMarketCap API: %s", err)
			return
		}
		coinDataCh <- coinData
	}()

	go func() {
		defer wg.Done()
		binanceData, err := fetchBinanceData(ctx)
		if err != nil {
			errCh <- fmt.Errorf("Error fetching data from Binance API: %s", err)
			return
		}
		binanceDataCh <- binanceData
	}()

	wg.Wait()
	close(coinDataCh)
	close(binanceDataCh)
	close(errCh)

	if len(errCh) > 0 {
		err := <-errCh
		fmt.Println("Error occurred:", err)
		return
	}

	coinData := <-coinDataCh
	binanceData := <-binanceDataCh

	if err := writeDataToCSV(coinData, binanceData, "btc_prices.csv"); err != nil {
		fmt.Println("Error writing data to CSV:", err)
	}
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
	// create a context with cancel func
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
	defer close(sigCh)
	// Create a ticker that trigger every 5 minutes
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	// invoke the initial function
	go fetchAndWriteDataToCSV(ctx, cancel)

	// Run script every 5 minutes
	for {
		select {
		case <-ticker.C:
			go fetchAndWriteDataToCSV(ctx, cancel)
		case <-sigCh:
			fmt.Println("Received terminiation signal")
			cancel()
			return
		case <-ctx.Done():
			fmt.Println("Exiting...")
			return
		}
	}

}
