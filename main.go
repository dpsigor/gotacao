package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/dpsigor/hltrnty"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

var defaultTickers []string = []string{
	"AAPL34",
	"B3SA3",
	"BBDC4",
	"GOGL34",
	"ITUB4",
	"NVDC34",
	"TSLA34",
	"CVCB3",
	"FBOK34",
	"IVVB11",
}

type Cotacao struct {
	Ticker    string `json:"ticker"`
	Min       string `json:"min"`
	Price     string `json:"price"`
	Max       string `json:"max"`
	PrevClose string `json:"prev_close"`
}

var r = regexp.MustCompile(`(?s)YMlKec.+?>R\$(.+?)<.+?last closing price<.+?P6K39c">R\$(.+?)<.+?P6K39c">R\$(.+?)<`)
var rminmax = regexp.MustCompile(`R\$`)

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

func queryTicker(v string) Cotacao {
	bhtml, err := reqHTML(v)
	if err != nil {
		fmt.Println(err)
		return Cotacao{}
	}
	c, err := parseHTML(v, bhtml)
	if err != nil {
		fmt.Println(err)
		return Cotacao{}
	}
	return c
}

func makeRow(c Cotacao) table.Row {
	prc, err := strconv.ParseFloat(c.Price, 32)
	if err != nil {
		prc = 0
	}
	prv, err := strconv.ParseFloat(c.PrevClose, 32)
	if err != nil {
		prv = 0
	}
	variacao := "0"
	if prc != 0 && prv != 0 {
		v := (prc - prv) * 100 / prv
		if v < 0 {
			variacao = fmt.Sprintf("\x1b[31m%v%%\x1b[0m", v)
		} else {
			variacao = fmt.Sprintf("%v%%", v)
		}
	}
	row := table.Row{
		c.Ticker,
		fmt.Sprintf("%8v", c.Price),
		fmt.Sprintf("%8v", c.Min),
		fmt.Sprintf("%8v", c.Max),
		fmt.Sprintf("%8v", c.PrevClose),
		variacao,
	}
	return row
}

func outputTable(cotacoes []Cotacao) {
	rows := hltrnty.Map(cotacoes, makeRow)
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	headers := table.Row{"Ticker", "Price", "Min", "Max", "PrevClose"}
	t.AppendHeader(headers)
	t.AppendRows(rows)
	t.SetStyle(table.Style{
		Color: table.ColorOptions{
			Header:       text.Colors{text.BgBlue},
			Row:          text.Colors{text.BgGreen, text.FgBlack},
			RowAlternate: text.Colors{text.BgYellow, text.FgBlack},
		},
		Box: table.BoxStyle{
			PaddingLeft: "  ",
		},
	})
	fmt.Println("")
	t.Render()
	fmt.Println("")
}

func main() {

	tojson := flag.Bool("j", false, "Outputs to json")
	dotime := flag.Bool("t", false, "Outputs how long the command took to run")
	flag.Parse()
	args := flag.Args()

	if len(args) > 100 {
		log.Fatal("Máximo 100 por vez")
	}

	tickers := defaultTickers
	if len(args) > 0 {
		tickers = hltrnty.Map(args, strings.ToUpper)
	}
	start := time.Now()

	cotacoes := hltrnty.ConcurMap(tickers, queryTicker)
	sort.Slice(cotacoes, func(a, b int) bool {
		return cotacoes[a].Ticker < cotacoes[b].Ticker
	})

	if !*tojson {
		outputTable(cotacoes)
	} else {
		err := json.NewEncoder(os.Stdout).Encode(cotacoes)
		if err != nil {
			log.Fatal(err)
		}
	}

	if *dotime {
		fmt.Printf("%s demorou %v para %v tickers\n", "gotacao", time.Since(start), len(tickers))
	}
}
