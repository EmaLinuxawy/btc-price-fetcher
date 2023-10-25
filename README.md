# CoinMarketCap Bitcoin Price Fetcher

This Go script allows you to fetch Bitcoin (BTC) prices from CoinMarketCap using an API token provided as the `COIN_MARKET_CAP_TOKEN` environment variable.

## Prerequisites

Before you can use this script, you need to obtain an API token from CoinMarketCap.

1. Visit the [CoinMarketCap Pro API](https://pro.coinmarketcap.com) website.
2. Sign up or log in to your account.
3. Generate an API token by following their documentation.

## Installation

1. Clone this repository or download the Go script to your local machine.

   ```bash
   git clone git@github.com:EmaLinuxawy/btc-price-fetcher.git
   cd btc-price-fetcher
   ```
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
Date: 2023-10-25 05:13 BTC Price: $ 34166.28
```