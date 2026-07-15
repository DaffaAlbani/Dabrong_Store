package main

import (
	"fmt"
	"ml-topup-v2/tokovoucher"
)

func main() {
	products, err := tokovoucher.GetAllProducts()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	
	fmt.Printf("Total Tokovoucher products: %d\n", len(products))
	
	for _, p := range products {
		code, _ := p["code"].(string)
		if code == "FF5" || code == "MFF5" || code == "ML3" || code == "SML3" {
			fmt.Printf("Found code %s: price=%v, name=%v\n", code, p["price"], p["nama_produk"])
		}
	}
}
