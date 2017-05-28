package main

import (
	"fmt"

	"github.com/DistributedSolutions/LendingBot/app/jsonrpc"
)

type HexInput struct {
	Hex string `json:"hex"`
}

type StringResp struct {
	Message string `json:"message"`
}

func main() {
	h := new(HexInput)
	h.Hex = "0000000000"

	sr := new(StringResp)
	req := jsonrpc.NewJSONRPCRequest("load-day", h, 0)
	_, jErr, err := req.POSTRequest("http://localhost:8080/api", sr)
	if err != nil {
		fmt.Println("ERROR", err)
	}

	if jErr != nil {
		fmt.Println("ERROR", jErr.Message)
	}
	fmt.Println(sr)
}
