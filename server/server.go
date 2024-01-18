package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type Quotation struct {
	CodeOut   string    `json:"code_out"`
	CodeIn    string    `json:"code_in"`
	Bid       float32   `json:"bid"`
	Timestamp time.Time `json:"timestamp"`
}

type QuotationResponse struct {
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

func QuotationMapper(q QuotationResponse) Quotation {
	value, err := strconv.ParseFloat(q.Usdbrl.Bid, 32)
	if err != nil {
		panic(err)
	}
	timestamp, err := strconv.ParseInt(q.Usdbrl.Timestamp, 10, 64)
	if err != nil {
		panic(err)
	}

	return Quotation{
		CodeOut:   q.Usdbrl.Code,
		CodeIn:    q.Usdbrl.Codein,
		Bid:       float32(value),
		Timestamp: time.Unix(timestamp, 0),
	}
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/cotacao", QuotationHandler)
	http.ListenAndServe(":8080", mux)
}

func QuotationHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	select {
	case <-r.Context().Done():
		log.Println("Request encerrada pelo cliente")
		// w.WriteHeader(http.StatusInternalServerError)
		// return
	case <-ctx.Done():
		log.Println("Tempo de 200ms excedido")
		w.WriteHeader(http.StatusGatewayTimeout)
		return
	default:
		q, err := fetchQuotationAPI(ctx)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		bStr := strconv.FormatFloat(float64(q.Bid), 'f', -1, 32)

		w.Write([]byte(bStr))
		registerQuotation(q)
		return
	}
}

func fetchQuotationAPI(ctx context.Context) (Quotation, error) {
	// ctx, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
	// defer cancel()
	log.Println("Requisição à API de cotação iniciada")
	defer log.Println("Requisição à API de cotação finalizada")
	req, err := http.Get("https://economia.awesomeapi.com.br/json/last/USD-BRL")
	if err != nil {
		return Quotation{}, err
	}
	res, err := io.ReadAll(req.Body)
	if err != nil {
		return Quotation{}, err
	}
	defer req.Body.Close()

	var qr QuotationResponse
	err = json.Unmarshal(res, &qr)
	if err != nil {
		return Quotation{}, err
	}
	q := QuotationMapper(qr)
	log.Println("Cotação obtida com sucesso", q)
	return q, nil
}

func registerQuotation(q Quotation) {
	// ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	// defer cancel()
	db, err := sql.Open("mysql", "db_user:pw123@tcp(localhost:3306)/exchange")
	if err != nil {
		panic(err)
	}
	err = insertQuotation(db, q)
	if err != nil {
		panic(err)
	}
	defer db.Close()
}

func insertQuotation(db *sql.DB, q Quotation) error {
	stmt, err := db.Prepare("insert into quotations(code_out,code_in,bid,timestamp) values(?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(q.CodeOut, q.CodeIn, q.Bid, q.Timestamp)
	if err != nil {
		return err
	}
	return nil
}
