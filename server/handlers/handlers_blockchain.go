package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
)

func (a *Api) HandleGetTransactions() echo.HandlerFunc {
	type input struct {
		Cursor  string
		Address string `validate:"required,eth_addr"`
		Chain   string `validate:"required"`
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
		Cursor       string        `json:"cursor"`
		Transactions []Transaction `json:"transactions"`
		Total        int64         `json:"total"`
	}

	return func(c echo.Context) error {
		in := input{
			Cursor:  c.QueryParam("cursor"),
			Address: c.Param("address"),
			Chain:   c.QueryParam("chain"),
		}
		if err := c.Validate(&in); err != nil {
			return err
		}

		moralisUrl := fmt.Sprintf("https://deep-index.moralis.io/api/v2/%s?chain=%s&limit=10", in.Address, in.Chain)
		if in.Cursor == "" {
			moralisUrl += "&disable_total=false"
		} else {
			moralisUrl = fmt.Sprintf("%s&disable_total=true&cursor=%s", moralisUrl, in.Cursor)
		}

		req, err := http.NewRequest("GET", moralisUrl, nil)
		if err != nil {
			return fmt.Errorf("failed to instantiate request: %w", err)
		}
		req.Header.Add("accept", "application/json")
		req.Header.Add("X-API-Key", a.cfg.MoralisApiKey)

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
			var response moralisResponse
			if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
				return fmt.Errorf("failed to decode moralis response: %w", err)
			}
			out := output{
				Cursor:       response.Cursor,
				Transactions: response.Result,
				Total:        response.Total,
			}

			return c.JSON(http.StatusOK, out)
		default:
			var response ErrorResponse
			if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
				return fmt.Errorf("failed to decode moralis response: %w", err)
			}
			return echo.NewHTTPError(http.StatusBadRequest, response.Message)
		}
	}
}

func (a *Api) HandleGetBalance() echo.HandlerFunc {
	type input struct {
		Address string `validate:"required,eth_addr"`
		Chain   string `validate:"required"`
	}

	type output struct {
		Balance string `json:"balance"`
	}
	return func(c echo.Context) error {
		in := input{
			Address: c.Param("address"),
			Chain:   c.QueryParam("chain"),
		}
		if err := c.Validate(&in); err != nil {
			return err
		}

		moralisUrl := fmt.Sprintf("https://deep-index.moralis.io/api/v2/%s/balance?chain=%s", in.Address, in.Chain)
		req, err := http.NewRequest("GET", moralisUrl, nil)
		if err != nil {
			return fmt.Errorf("failed to instantiate request: %w", err)
		}
		req.Header.Add("accept", "application/json")
		req.Header.Add("X-API-Key", a.cfg.MoralisApiKey)

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
			var response output
			if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
				return fmt.Errorf("failed to decode moralis response: %w", err)
			}
			return c.JSON(http.StatusOK, response)
		default:
			var response ErrorResponse
			if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
				return fmt.Errorf("failed to decode moralis response: %w", err)
			}
			return echo.NewHTTPError(http.StatusBadRequest, response.Message)
		}
	}
}
