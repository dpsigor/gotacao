package main

import (
	"fmt"
	"testing"
)

func TestReqHTML(t *testing.T) {
	bhtml, err := reqHTML("ITUB4")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(bhtml))
}

func TestQueryTicker(t *testing.T) {
	c, err := queryTicker("ITUB4")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%+v\n", c)
}
