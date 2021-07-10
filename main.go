package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/jedib0t/go-pretty/v6/table"
)

var rminmax *regexp.Regexp = regexp.MustCompile(`R\$`)

func reqCotacao(ticker string, r *regexp.Regexp) (table.Row, error) {
	resp, err := http.Get("https://www.google.com/finance/quote/" + ticker + ":BVMF")
	if err != nil {
		log.Print(err)
		return nil, err
	}
	b, _ := ioutil.ReadAll(resp.Body)
	m := r.FindSubmatch(b)

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

	row := table.Row{ticker, min, price, max, prevClose}
	return row, nil
}

func main() {
	r := regexp.MustCompile(`YMlKec fxKbKc[\w\W\n]+?>R\$(?P<price>.+?)<[\w\W\n]+?last closing price<[\w\W\n]+?P6K39c">R\$(?P<prevClose>.+?)<[\w\W\n]+?P6K39c">R\$(?P<minmax>.+?)<`)
	// petr4 itub4 b3sa3 bkbr3 aapl34 tsla34 amzo34 gogl34 cmig4 bbas3 bbdc4
	tickers := []string{
		"AAPL34",
		"AMZO34",
		"BBDC4",
		"GOGL34",
		"ITUB4",
		"NVDC34",
		"TSLA34",
	}
	args := os.Args[1:]

	if len(args) > 100 {
		log.Fatal("Máximo 100 por vez")
	}

	if len(args) > 0 {
		tickers = []string{}
		for _, v := range args {
			tickers = append(tickers, strings.ToUpper(v))
		}
	}

	var wg sync.WaitGroup

	var rows []table.Row

	for _, v := range tickers {
		wg.Add(1)
		go func(v string) {
			defer wg.Done()
			row, err := reqCotacao(v, r)
			if err != nil {
				fmt.Println(err)
				return
			}
			rows = append(rows, row)
		}(v)
	}

	wg.Wait()
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	headers := table.Row{"Ticker", "Min", "Price", "Max", "PrevClose"}
	t.AppendHeader(headers)
	t.AppendRows(rows)
	t.Render()
}
