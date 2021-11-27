package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
)

type Cotacao struct {
	Ticker    string `json:"ticker"`
	Min       string `json:"min"`
	Price     string `json:"price"`
	Max       string `json:"max"`
	PrevClose string `json:"prev_close"`
}

var r = regexp.MustCompile(`(?s)YMlKec.+?>R\$(.+?)<.+?last closing price<.+?P6K39c">R\$(.+?)<.+?P6K39c">R\$(.+?)<`)
var rminmax = regexp.MustCompile(`R\$`)

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

func idxof(s []string, str string) int {
	for i, v := range s {
		if v == str {
			return i
		}
	}
	return -1
}

func rmel(s []string, str string) []string {
	if len(s) > 1 {
		idx := idxof(s, str)
		s = append(s[:idx], s[idx+1:]...)
	} else {
		s = []string{}
	}
	return s
}

func hdlargs(args []string) ([]string, bool, bool, time.Time) {
	var start time.Time
	var tojson bool
	if len(args) > 0 && contains(args, "-j") {
		tojson = true
		args = rmel(args, "-j")
	}

	var dotime bool
	if len(args) > 0 && contains(args, "-t") {
		dotime = true
		start = time.Now()
		args = rmel(args, "-t")
	}
	return args, tojson, dotime, start
}

func parseHTML(ticker string, body []byte) (Cotacao, error) {
	m := r.FindSubmatch(body)

	var min string
	var price string
	var max string
	var prevClose string

	if len(m) > 1 {
		price = string(m[1])
	}

	if len(m) > 2 {
		prevClose = string(m[2])
	}

	if len(m) > 3 {
		bminmax := rminmax.ReplaceAll(m[3], []byte(""))
		minmax := string(bminmax)
		sminmax := strings.Split(minmax, "-")
		if len(sminmax) > 1 {
			min = sminmax[0]
			max = sminmax[1]
		}
	}

	cotacao := Cotacao{
		Ticker:    ticker,
		Min:       min,
		Price:     price,
		Max:       max,
		PrevClose: prevClose,
	}
	return cotacao, nil
}

func reqHTML(ticker string) ([]byte, error) {
	resp, err := http.Get("https://www.google.com/finance/quote/" + ticker + ":BVMF")
	if err != nil {
		log.Print(err)
		return []byte{}, err
	}
	b, _ := ioutil.ReadAll(resp.Body)
	return b, nil
}

func queryTicker(v string) (Cotacao, error) {
	bhtml, err := reqHTML(v)
	if err != nil {
		return Cotacao{}, err
	}
	c, err := parseHTML(v, bhtml)
	if err != nil {
		return Cotacao{}, err
	}
	return c, nil
}

func main() {
	// petr4 itub4 b3sa3 bkbr3 aapl34 tsla34 amzo34 gogl34 cmig4 bbas3 bbdc4
	tickers := []string{
		"AAPL34",
		"AMZO34",
		"BBDC4",
		"GOGL34",
		"ITUB4",
		"NVDC34",
		"TSLA34",
		"CVCB3",
		"FBOK34",
		"IVVB11",
	}
	args := os.Args[1:]

	args, tojson, dotime, start := hdlargs(args)

	if len(args) > 100 {
		log.Fatal("Máximo 100 por vez")
	}

	if len(args) > 0 {
		tickers = []string{}
		for _, v := range args {
			tickers = append(tickers, strings.ToUpper(v))
		}
	}

	var cotacoes []Cotacao

	var wg sync.WaitGroup
	for _, v := range tickers {
		wg.Add(1)
		go func(v string) {
			defer wg.Done()
			c, err := queryTicker(v)
			if err != nil {
				fmt.Println(err)
				return
			}
			cotacoes = append(cotacoes, c)
		}(v)
	}
	wg.Wait()

	sort.Slice(cotacoes, func(a, b int) bool {
		return cotacoes[a].Ticker < cotacoes[b].Ticker
	})

	if !tojson {
		var rows []table.Row
		for _, c := range cotacoes {
			row := table.Row{c.Ticker, c.Min, c.Price, c.Max, c.PrevClose}
			rows = append(rows, row)
		}
		t := table.NewWriter()
		t.SetOutputMirror(os.Stdout)
		headers := table.Row{"Ticker", "Min", "Price", "Max", "PrevClose"}
		t.AppendHeader(headers)
		t.AppendRows(rows)
		t.Render()
	} else {
		j, err := json.Marshal(cotacoes)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(j))
	}

	if dotime {
		fmt.Printf("%s demorou %v para %v tickers\n", "gotacao", time.Since(start), len(tickers))
	}
}
