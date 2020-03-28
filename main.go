package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
)

type Currencies struct {
	Id          string `json:"id,omitempty"`
	FullName    string `json:"fullName,omitempty"`
	Ask         string `json:"ask,omitempty"`
	Bid         string `json:"bid,omitempty"`
	Last        string `json:"last,omitempty"`
	Open        string `json:"open,omitempty"`
	Low         string `json:"low,omitempty"`
	High        string `json:"high,omitempty"`
	FeeCurrency string `json:"feeCurrency,omitempty"`
}

type CurrencyTicket struct {
	Ask         string `json:"ask"`
	Bid         string `json:"bid"`
	Last        string `json:"last"`
	Open        string `json:"open"`
	Low         string `json:"low"`
	High        string `json:"high"`
	Volume      string `json:"volume,omitempty"`
	VolumeQuote string `json:"volumeQuote,omitempty"`
	Timestamp   string `json:"timestamp,omitempty"`
	Symbol      string `json:"symbol"`
}

type CurrencySymbol struct {
	Id                   string `json:"id"`
	BaseCurrency         string `json:"baseCurrency"`
	QuoteCurrency        string `json:"quoteCurrency"`
	QuantityIncrement    string `json:"quantityIncrement,omitempty"`
	TickSize             string `json:"tickSize,omitempty"`
	TakeLiquidityRate    string `json:"takeLiquidityRate,omitempty"`
	ProvideLiquidityRate string `json:"provideLiquidityRate,omitempty"`
	FeeCurrency          string `json:"feeCurrency,omitempty"`
}

type CurrencyInfo struct {
	Id                 string `json:"id"`
	FullName           string `json:"fullName"`
	Crypto             bool   `json:"crypto,omitempty"`
	PayinEnabled       bool   `json:"payinEnabled,omitempty"`
	PayinPaymentId     bool   `json:"payinPaymentId,omitempty"`
	PayinConfirmations int    `json:"payinConfirmations,omitempty"`
	PayoutEnabled      bool   `json:"payoutEnabled,omitempty"`
	PayoutIsPaymentId  bool   `json:"payoutIsPaymentId,omitempty"`
	TransferEnabled    bool   `json:"transferEnabled,omitempty"`
	Delisted           bool   `json:"delisted,omitempty"`
	PayoutFee          string `json:"payoutFee,omitempty"`
}

var curr []Currencies
var wg = sync.WaitGroup{}

func main() {
	fmt.Println("Hello Welcome Golang")
	router := mux.NewRouter()

	curr = append(curr, Currencies{Id: "ETH", FullName: "Ethereum", Ask: "0.054464", Bid: "0.054463", Last: "0.054463", Open: "0.057133", Low: "0.053615", High: "0.057559", FeeCurrency: "BTC"})
	curr = append(curr, Currencies{Id: "BTC", FullName: "Bitcoin", Ask: "7906.72", Bid: "7906.28", Last: "7906.48", Open: "7952.3", Low: "7561.51", High: "8107.96", FeeCurrency: "USD"})

	router.HandleFunc("/currency/all", GetAllCurrencyEndpoint).Methods("GET")
	router.HandleFunc("/currency/{symbol}", GetSymbolCurrencyEndpoint).Methods("GET")
	log.Fatal(http.ListenAndServe(":12345", router))
}

func GetAllCurrencyEndpoint(w http.ResponseWriter, req *http.Request) {
	fmt.Println("GetEmployeeEndpoint called")
	resBody := getUrlResponce("https://api.hitbtc.com/api/2/public/symbol")
	currencySymbol := []CurrencySymbol{}
	if err := json.Unmarshal(resBody, &currencySymbol); err != nil {
		log.Println(err)
	}

	count := 1
	//currSymSize := unsafe.Sizeof(currencySymbol)
	for _, item := range currencySymbol {
		if count == 50 {
			break
		}

		count++
		wg.Add(1)
		go setCurrencyDetails(item)
	}
	wg.Wait()

	json.NewEncoder(w).Encode(curr)
}

func GetSymbolCurrencyEndpoint(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	for _, item := range curr {
		if item.Id == params["symbol"] {
			json.NewEncoder(w).Encode(item)
			return
		}
	}
}

func setCurrencyDetails(item CurrencySymbol) {
	var currInfo CurrencyInfo
	resBody := getUrlResponce("https://api.hitbtc.com/api/2/public/currency/" + item.BaseCurrency)
	if err := json.Unmarshal(resBody, &currInfo); err != nil {
		fmt.Println("CurrencyInfo err msg came ", err)
		wg.Done()
		return
	}

	var eachReq Currencies

	eachReq.Id = item.BaseCurrency
	eachReq.FeeCurrency = item.FeeCurrency
	eachReq.FullName = currInfo.FullName

	var currTick CurrencyTicket
	tickBody := getUrlResponce("https://api.hitbtc.com/api/2/public/ticker/" + item.Id)
	if err := json.Unmarshal(tickBody, &currTick); err != nil {
		fmt.Println("CurrencyTicket err msg came ", err)
		wg.Done()
		return
	}

	eachReq.Ask = currTick.Ask
	eachReq.Bid = currTick.Bid
	eachReq.Last = currTick.Last
	eachReq.Open = currTick.Open
	eachReq.Low = currTick.Low
	eachReq.High = currTick.High

	curr = append(curr, eachReq)
	wg.Done()
}

func getUrlResponce(requestURL string) []byte {
	response, err := http.Get(requestURL)
	if err != nil {
		fmt.Print(err.Error())
		return nil
	}

	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Print(err.Error())
		return nil
	}
	return responseData
}
