/*
- O server.go deverá consumir a API contendo o câmbio de Dólar e Real no endereço: https://economia.awesomeapi.com.br/json/last/USD-BRL e em seguida deverá retornar no formato JSON o resultado para o cliente.
- Usando o package "context", o server.go deverá registrar no banco de dados SQLite cada cotação recebida, sendo que o timeout máximo para chamar a API de cotação do dólar deverá ser de 200ms e o timeout máximo para conseguir persistir os dados no banco deverá ser de 10ms.
- O endpoint necessário gerado pelo server.go para este desafio será: /cotacao e a porta a ser utilizada pelo servidor HTTP será a 8080.
- Os contextos deverão retornar erro nos logs caso o tempo de execução seja insuficiente.
*/
package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const quotationApiUrl string = "https://economia.awesomeapi.com.br/json/last/USD-BRL"
const apiRequestTimeout time.Duration = 200 * time.Millisecond
const dbSaveTimeout time.Duration = 10 * time.Millisecond
const fileDb string = "./quotations.db"
const createTable string = `
  CREATE TABLE IF NOT EXISTS quotations (
  ID INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
  Code       TEXT NOT NULL,
  Codein     TEXT NOT NULL,
  Name       TEXT NOT NULL,
  High       FLOAT NOT NULL,
  Low        FLOAT NOT NULL,
  VarBid     FLOAT NOT NULL,
  PctChange  FLOAT NOT NULL,
  Bid        FLOAT NOT NULL,
  Ask        FLOAT NOT NULL,
  Timestamp  TIMESTAMP NOT NULL,
  CreateDate DATETIME NOT NULL
  );`
const insertQuery = `
	INSERT INTO quotations (Code, Codein, Name, High, Low, VarBid, PctChange, Bid, Ask, Timestamp, CreateDate) 
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`

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
	initDatabase()
	initWebServer()
}

func initDatabase() {
	_, err := os.Stat(fileDb)
	if os.IsNotExist(err) {
		_, err := os.Create(fileDb)
		checkError(err)
	}
	createQuotationTable()
}

func createQuotationTable() {
	db, err := openDatabase()
	checkError(err)
	defer db.Close()

	_, err = db.Exec(createTable)
	checkError(err)
}

func initWebServer() {
	http.HandleFunc("/cotacao", getQuotationHandler)
	http.ListenAndServe(":8080", nil)
}

func openDatabase() (*sql.DB, error) {
	return sql.Open("sqlite3", fileDb)
}

func getQuotationHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Request init")
	defer log.Println("Request end")

	ctx, cancel := context.WithTimeout(context.Background(), apiRequestTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", quotationApiUrl, nil)
	checkError(err)

	res, err := http.DefaultClient.Do(req)
	checkError(err)
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	checkError(err)

	var quotation Quotation
	err = json.Unmarshal(body, &quotation)
	checkError(err)

	saveQuotation(quotation)

	w.Write(body)
}

func saveQuotation(c Quotation) {
	db, err := openDatabase()
	checkError(err)
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), dbSaveTimeout)
	defer cancel()

	_, err = db.ExecContext(ctx, insertQuery, c.Usdbrl.Code, c.Usdbrl.Codein, c.Usdbrl.Name,
		c.Usdbrl.High, c.Usdbrl.Low, c.Usdbrl.VarBid, c.Usdbrl.PctChange, c.Usdbrl.Bid,
		c.Usdbrl.Ask, c.Usdbrl.Timestamp, c.Usdbrl.CreateDate)
	checkError(err)
}

func checkError(err error) {
	if err != nil {
		log.Panicln(err)
	}
}
