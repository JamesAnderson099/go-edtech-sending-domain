package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"example.com/edtech-domain-onboarding/domainonboard"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: domain-onboard <sending-domain>")
		os.Exit(2)
	}

	client := domainonboard.New(os.Getenv("INFRAI_API_KEY"))
	result, err := client.VerifyDomain(context.Background(), os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	output, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println(string(output))
}
