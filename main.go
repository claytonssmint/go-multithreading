package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type AddressBrasilAPI struct {
	Cep          string `json:"cep"`
	State        string `json:"state"`
	City         string `json:"city"`
	Neighborhood string `json:"neighborhood"`
	Street       string `json:"street"`
	Service      string `json:"service"`
}

type AddressViaCEP struct {
	Cep         string `json:"cep"`
	Logradouro  string `json:"logradouro"`
	Complemento string `json:"complemento"`
	Unidade     string `json:"unidade"`
	Bairro      string `json:"bairro"`
	Localidade  string `json:"localidade"`
	Uf          string `json:"uf"`
	Ibge        string `json:"ibge"`
	Gia         string `json:"gia"`
	Ddd         string `json:"ddd"`
	Siafi       string `json:"siafi"`
}

type Result struct {
	Address interface{}
	Source  string
	Error   error
}

func fetchingFromAPI(ctx context.Context, url string, source string, address interface{}, resultChan chan<- Result) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		resultChan <- Result{Error: err}
		return
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		resultChan <- Result{Error: err}
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		resultChan <- Result{Error: fmt.Errorf("unexpected status code: %d", resp.StatusCode)}
		return
	}
	if err := json.NewDecoder(resp.Body).Decode(&address); err != nil {
		resultChan <- Result{Error: err}
		return
	}

	resultChan <- Result{Address: address, Source: source}
}

func main() {
	cep := "05874120"

	brasilAPIURL := fmt.Sprintf("https://brasilapi.com.br/api/cep/v1/%s", cep)
	viaCEPAPIURL := fmt.Sprintf("https://viacep.com.br/ws/%s/json/", cep)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	resultChan := make(chan Result, 2)

	go fetchingFromAPI(ctx, brasilAPIURL, "BrasilAPI", &AddressBrasilAPI{}, resultChan)
	go fetchingFromAPI(ctx, viaCEPAPIURL, "ViaCEP", &AddressViaCEP{}, resultChan)

	select {
	case result := <-resultChan:
		if result.Error != nil {
			fmt.Println(result.Error)
		} else {
			fmt.Printf("Source: %s\nAddress: %+v\n", result.Source, result.Address)
		}
	case <-ctx.Done():
		fmt.Println("Error: timeout")
	}
}
