package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
)

func (a *Api) HandleGetTransactions() echo.HandlerFunc {
	type input struct {
		Address string `param:"address" validate:"required,ethereum_address"`
		Chain   string `query:"chain" validate:"required"`
	}

	type Transaction struct {
		Hash                     string  `json:"hash"`
		Nonce                    string  `json:"nonce"`
		TransactionIndex         string  `json:"transaction_index"`
		FromAddress              string  `json:"from_address"`
		ToAddress                string  `json:"to_address"`
		Value                    string  `json:"value"`
		Gas                      string  `json:"gas"`
		GasPrice                 string  `json:"gas_price"`
		Input                    string  `json:"input"`
		ReceiptCumulativeGasUsed string  `json:"receipt_cumulative_gas_used"`
		ReceiptGasUsed           string  `json:"receipt_gas_used"`
		ReceiptContractAddress   string  `json:"receipt_contract_address"`
		ReceiptRoot              string  `json:"receipt_root"`
		ReceiptStatus            string  `json:"receipt_status"`
		BlockTimestamp           string  `json:"block_timestamp"`
		BlockNumber              string  `json:"block_number"`
		BlockHash                string  `json:"block_hash"`
		TransferIndex            []int64 `json:"transfer_index"`
	}

	type moralisResponse struct {
		Total    int64         `json:"total"`
		PageSize int64         `json:"page_size"`
		Page     int64         `json:"page"`
		Cursor   string        `json:"cursor"`
		Result   []Transaction `json:"result"`
	}

	type output struct {
		Transactions []Transaction `json:"transactions"`
		Total        int64         `json:"total"`
	}

	return func(c echo.Context) error {
		var in input
		if err := c.Bind(&in); err != nil {
			return err
		}
		if err := c.Validate(&in); err != nil {
			return err
		}

		moralisUrl := fmt.Sprintf("https://deep-index.moralis.io/api/v2/%s?chain=%s&limit=50&disable_total=true", in.Address, in.Chain)

		var response moralisResponse
		err := fetchFromMoralis(moralisUrl, a.cfg.MoralisApiKey, &response)
		if err != nil {
			return err
		}

		out := output{
			Transactions: response.Result,
			Total:        response.Total,
		}

		return c.JSON(http.StatusOK, out)
	}
}

func (a *Api) HandleGetBalance() echo.HandlerFunc {
	type input struct {
		Address string `param:"address" validate:"required,ethereum_address"`
		Chain   string `query:"chain" validate:"required"`
	}

	type output struct {
		Balance string `json:"balance"`
	}
	return func(c echo.Context) error {
		var in input
		if err := c.Bind(&in); err != nil {
			return err
		}
		if err := c.Validate(&in); err != nil {
			return err
		}

		moralisUrl := fmt.Sprintf("https://deep-index.moralis.io/api/v2/%s/balance?chain=%s", in.Address, in.Chain)

		var out output
		err := fetchFromMoralis(moralisUrl, a.cfg.MoralisApiKey, &out)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, out)
	}
}

func fetchFromMoralis(url, apiKey string, out any) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to instantiate request: %w", err)
	}
	req.Header.Add("accept", "application/json")
	req.Header.Add("X-API-Key", apiKey)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer func() {
		_ = res.Body.Close()
	}()

	switch res.StatusCode {
	case http.StatusBadRequest:
		var response ErrorResponse
		if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
			return fmt.Errorf("failed to decode moralis response: %w", err)
		}
		return echo.NewHTTPError(http.StatusBadRequest, response.Message)
	case http.StatusOK:
		if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
			return fmt.Errorf("failed to decode moralis response: %w", err)
		}
		return nil
	default:
		var response ErrorResponse
		if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
			return fmt.Errorf("failed to decode moralis response: %w", err)
		}
		return echo.NewHTTPError(http.StatusBadRequest, response.Message)
	}
}
