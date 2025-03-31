package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
)

// Struct for the POST request sent to rounder sevice
type InputRequestBody struct {
	Transactions []InputTransaction `json:"results"`
	AccountCode  string             `json:"account_code"`
}

// Struct for each transaction in request
type InputTransaction struct {
	AccountCode string  `json:"account_code"`
	Timestamp   string  `json:"timestamp"`
	Description string  `json:"description"`
	Type        string  `json:"transaction_type"`
	Category    string  `json:"transaction_category"`
	Amount      float64 `json:"amount"`
	Currency    string  `json:"currency"`
}
type OutputTransaction struct {
	AccountCode   string  `json:"account_code"`
	Description   string  `json:"description"`
	RawAmount     float64 `json:"raw_amount"`
	RoundedAmount int     `json:"rounded_amount"`
	Delta         float64 `json:"delta"`
	Timestamp     string  `json:"timestamp"`
}

func pushTransactions(transactions []OutputTransaction) error {

	jsonData, err := json.Marshal(transactions)
	if err != nil {
		return fmt.Errorf("error marshaling transactions: %w", err)
	}

	// Create a POST request
	req, err := http.NewRequest("POST", "http://localhost:3001/api/transactions", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	// Check for a successful response
	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("server returned non-200 status: %s", resp.Status)
	}

	fmt.Println("Transactions sent successfully!")
	return nil
}

func roundTransactions(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Request recieved")
	var inputTransactions []InputTransaction
	var transactions []OutputTransaction
	totalToPay := 0.0

	body, _ := io.ReadAll(r.Body)
	err := json.Unmarshal([]byte(body), &inputTransactions)
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, transaction := range inputTransactions {
		var outputTransaction OutputTransaction
		roundedAmount := math.Ceil(transaction.Amount)
		deltaFloat := roundedAmount - transaction.Amount
		deltaFloat = math.Round(deltaFloat*100) / 100
		totalToPay += deltaFloat

		outputTransaction.Delta = deltaFloat
		outputTransaction.Description = transaction.Description
		outputTransaction.RawAmount = transaction.Amount
		outputTransaction.RoundedAmount = int(roundedAmount)
		outputTransaction.AccountCode = transaction.AccountCode
		outputTransaction.Timestamp = transaction.Timestamp

		transactions = append(transactions, outputTransaction)

	}

	err = pushTransactions(transactions)
	if err != nil{
		fmt.Println(err)
	}
}

func main() {
	// Listen to "new_transactions" topic
	// Get the amount and use Math.ceil to round up
	// Push to "rounded_transactions" topic UserId, transaction_amount, rounded_transaction_amount, savings
	fmt.Println("Listening")
	http.HandleFunc("POST /api/round-up", roundTransactions)
	http.ListenAndServe(":3000", nil)

}
