package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/mattn/go-sqlite3"
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

const dbFilePath = "./exchange.db"

func main() {
	CreateDatabaseAndTable(dbFilePath)
	mux := http.NewServeMux()
	mux.HandleFunc("/cotacao", QuotationHandler)
	http.ListenAndServe(":8080", mux)
}

func CreateDatabaseAndTable(filename string) {
	os.Remove(filename)
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	db, err := sql.Open("sqlite3", dbFilePath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	sqlStmt := `
    CREATE TABLE IF NOT EXISTS quotations (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        code_out TEXT NOT NULL,
        code_in TEXT NOT NULL,
        bid REAL NOT NULL,
        timestamp INTEGER NOT NULL
    );
    `
	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return
	}
}

func QuotationHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	select {
	case <-r.Context().Done():
		log.Println("Request encerrada pelo cliente")
	case <-ctx.Done():
		if ctx.Err() == context.DeadlineExceeded {
			log.Println("Tempo de 200ms excedido durante a inserção no banco de dados")
			w.WriteHeader(http.StatusGatewayTimeout)
			return
		}
		return
	default:
		q, err := fetchQuotationAPI(ctx)
		if err != nil {
			if ctx.Err() == context.DeadlineExceeded {
				log.Println("Tempo de 200ms excedido")
				return
			}
			panic(err)
		}
		bStr := strconv.FormatFloat(float64(q.Bid), 'f', -1, 32)

		w.Write([]byte(bStr))
		registerQuotation(q)
		return
	}
}

func fetchQuotationAPI(ctx context.Context) (Quotation, error) {
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
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	select {
	case <-ctx.Done():
		return
	default:
		db, err := sql.Open("sqlite3", dbFilePath)
		if err != nil {
			panic(err)
		}
		err = insertQuotation(db, q, ctx)
		if err != nil {
			if ctx.Err() == context.DeadlineExceeded {
				log.Println("Tempo de 200ms excedido")
				return
			}
			panic(err)
		}
		defer db.Close()
		log.Println("Cotação registrada no banco de dados com sucesso")
	}

}

func insertQuotation(db *sql.DB, q Quotation, ctx context.Context) error {
	stmt, err := db.Prepare("insert into quotations(code_out,code_in,bid,timestamp) values(?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.ExecContext(ctx, q.CodeOut, q.CodeIn, q.Bid, q.Timestamp.Unix())
	if err != nil {
		return err
	}
	return nil
}
