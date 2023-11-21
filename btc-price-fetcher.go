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
	"strconv"
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

// Global HTTP client to reuse for all HTTP requests.
var httpClient = &http.Client{}

// fetchData retrieves data from CoinMarketCap.
func fetchData(ctx context.Context) (CoinData, error) {
	uri := "https://pro-api.coinmarketcap.com/v1/cryptocurrency/listings/latest"

	// Check context before making a request
	if err := ctx.Err(); err != nil {
		return CoinData{}, err
	}

	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return CoinData{}, fmt.Errorf("Error creating HTTP request: %v", err)
	}

	q := url.Values{}
	q.Add("start", "1")
	q.Add("limit", "1")
	q.Add("convert", "USD")
	req.Header.Set("Accepts", "application/json")
	req.Header.Add("X-CMC_PRO_API_KEY", os.Getenv("COIN_MARKET_CAP_TOKEN"))
	req.URL.RawQuery = q.Encode()

	// Execute HTTP request with context
	resp, err := httpClient.Do(req.WithContext(ctx))
	if err != nil {
		return CoinData{}, fmt.Errorf("Error sending request to CoinMarketCap %v", err)
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

	return coinData, nil
}

// fetchBinanceData retrieves data from Binance API.
func fetchBinanceData(ctx context.Context) (BinanceData, error) {
	uri := "https://api.binance.com/api/v3/ticker/price?symbol=BTCUSDT"

	// Check context before making a request
	if err := ctx.Err(); err != nil {
		return BinanceData{}, err
	}

	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return BinanceData{}, fmt.Errorf("Error creating HTTP request: %w", err)
	}

	// Execute HTTP request with context
	resp, err := httpClient.Do(req.WithContext(ctx))
	if err != nil {
		return BinanceData{}, fmt.Errorf("Error sending HTTP request to BinanceAPI: %w", err)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return BinanceData{}, fmt.Errorf("error reading response body: %w", err)
	}

	var coinData BinanceData
	err = json.Unmarshal(respBody, &coinData)
	if err != nil {
		return BinanceData{}, fmt.Errorf("Error unmarshalling JSON from BinanceAPI: %w", err)
	}

	return coinData, nil
}

// fetchAndWriteDataToCSV fetches data and writes it to a CSV file.
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
		return
	}
}

// writeDataToCSV writes the given data to a CSV file.
func writeDataToCSV(data CoinData, binData BinanceData, filename string) error {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Can't open the file", err)
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	columnName := []string{"Date", "CoinMarketCap Price", "Binance Price"}
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to retrieve file statistics for %s: %w", filename, err)
	}

	// Check if the file is empty and write column names if so
	isEmpty := fileInfo.Size() == 0
	if isEmpty {
		if err := writer.Write(columnName); err != nil {
			return fmt.Errorf("error writing columns: %w", err)
		}
	}

	t := time.Now()
	timeNow := t.Format("2006-01-02 3:4:5 pm")

	// Convert Binance price from string to Float to format it
	binPrice, err := strconv.ParseFloat(binData.Price, 64)
	if err != nil {
		return fmt.Errorf("Error parsing Binance Price: %w", err)
	}
	formatedPrice := fmt.Sprintf("%.2f", binPrice)
	for _, btc := range data.Data[0].Quote {
		price := fmt.Sprintf("%.2f", btc.Price)
		if err := writer.Write([]string{timeNow, price, formatedPrice}); err != nil {
			return fmt.Errorf("error writing data row: %w", err)
		}
		fmt.Println("Date:", timeNow, "BTC Price: $", price)
	}
	fmt.Println("price of binance:", formatedPrice)

	return nil
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
	defer close(sigCh)
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	go fetchAndWriteDataToCSV(ctx, cancel)

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
