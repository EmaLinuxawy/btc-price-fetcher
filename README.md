# Bitcoin Price Tracker - CoinMarketCap & Binance

Efficiently track Bitcoin prices with this Go script, fetching real-time data from CoinMarketCap and Binance. Utilizes context for optimized request handling and signal processing for graceful shutdowns. Outputs are conveniently saved to a CSV file, making data tracking and analysis straightforward.

Perfect for cryptocurrency enthusiasts and developers interested in real-time Bitcoin price monitoring and data aggregation.

This Go script allows you to fetch Bitcoin (BTC) prices from both CoinMarketCap and Binance. It uses the CoinMarketCap API token provided as the `COIN_MARKET_CAP_TOKEN` environment variable and accesses the Binance API without a token.

## Prerequisites

Before using this script, you need to obtain an API token from CoinMarketCap.

1. Visit the [CoinMarketCap Pro API](https://pro.coinmarketcap.com) website.
2. Sign up or log in to your account.
3. Generate an API token by following their documentation.

## Features

- Fetch BTC prices from CoinMarketCap and Binance APIs.
- Uses context to manage and cancel HTTP requests effectively.
- Handles OS signals for graceful termination of the script.
- Saves fetched data to a CSV file for easy tracking.

## Installation

1. Clone this repository or download the Go script to your local machine.

   ```bash
   git clone git@github.com:EmaLinuxawy/btc-price-fetcher.git
   cd btc-price-fetcher

2. Build the Go script

   ```bash
   go build
   ```

## Usage

1. Set the COIN_MARKET_CAP_TOKEN environment variable to your CoinMarketCap API token.

   ```bash
   export COIN_MARKET_CAP_TOKEN="YOUR_API_TOKEN"
2. Run the script to fetch the BTC Price and save it to the CSV

   ```bash
   ./btc-price-fetcher
## Example Output

```bash
Date: 2023-11-20 9:33:52 pm BTC Price: $ 37564.62
price of binance: 37536.44
```
