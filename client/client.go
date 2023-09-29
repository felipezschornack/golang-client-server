/*
- O client.go deverá realizar uma requisição HTTP no server.go solicitando a cotação do dólar.
- O client.go precisará receber do server.go apenas o valor atual do câmbio (campo "bid" do JSON). Utilizando o package "context", o client.go terá um timeout máximo de 300ms para receber o resultado do server.go.
- O client.go terá que salvar a cotação atual em um arquivo "cotacao.txt" no formato: Dólar: {valor}
- O endpoint necessário gerado pelo server.go para este desafio será: /cotacao e a porta a ser utilizada pelo servidor HTTP será a 8080.
- Os contextos deverão retornar erro nos logs caso o tempo de execução seja insuficiente.
*/
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

const resultFileName string = "cotacao.txt"
const requestTimeout time.Duration = 300 * time.Millisecond
const apiUrl string = "http://localhost:8080/cotacao"

type Quotation struct {
	Usdbrl struct {
		Bid string `json:"bid"`
	} `json:"USDBRL"`
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", apiUrl, nil)
	checkError(err)

	res, err := http.DefaultClient.Do(req)
	checkError(err)
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	checkError(err)

	var quotation Quotation
	err = json.Unmarshal(body, &quotation)
	checkError(err)

	saveInFile(quotation)
}

func saveInFile(q Quotation) {
	f, err := os.Create(resultFileName)
	checkError(err)
	defer f.Close()

	_, err = f.WriteString(fmt.Sprintf("Dólar: {%s}", q.Usdbrl.Bid))
	checkError(err)
}

func checkError(err error) {
	if err != nil {
		log.Panicln(err)
	}
}
