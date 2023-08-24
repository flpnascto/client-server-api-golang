package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

type Quotation struct {
	Usdbrl struct {
		Code       string `json:"code"`
		Codein     string `json:"codein"`
		Name       string `json:"name"`
		High       string `json:"high"`
		Low        string `json:"low"`
		VarBid     string `json:"varBid"`
		PctChange  string `json:"pctChange"`
		Bid        string `json:"bid"`
		Ask        string `json:"ask"`
		Timestamp  string `json:"timestamp"`
		CreateDate string `json:"create_date"`
	} `json:"USDBRL"`
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/cotacao", QuotationHandler)
	http.ListenAndServe(":8080", mux)
}

func QuotationHandler(w http.ResponseWriter, r *http.Request) {
	bid := fetchQuotationAPI()
	w.Write([]byte(bid))
	bidFloat, error := strconv.ParseFloat(bid, 32)
	if error != nil {
		panic(error)
	}
	registerQuotation(float32(bidFloat))
}

func fetchQuotationAPI() string {
	resp, error := http.Get("https://economia.awesomeapi.com.br/json/last/USD-BRL")
	if error != nil {
		panic(error)
	}
	defer resp.Body.Close()

	body, error := io.ReadAll(resp.Body)
	if error != nil {
		panic(error)
	}

	var q Quotation
	error = json.Unmarshal(body, &q)
	if error != nil {
		panic(error)
	}
	b := q.Usdbrl.Bid
	return b
}

func registerQuotation(bid float32) {
	fmt.Printf("Implementar o registro da cotação R$%.2f no banco de dados SQLite", bid)
}
