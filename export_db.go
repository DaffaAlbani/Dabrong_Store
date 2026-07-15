package main

import (
	"encoding/json"
	"fmt"
	"ml-topup-v2/database"
	"os"
)

func main() {
	database.InitDB("./orders.db")
	products, err := database.GetCachedProducts()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	data, err := json.MarshalIndent(products, "", "  ")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	err = os.WriteFile("database/products.json", data, 0644)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("Successfully exported products to database/products.json")
}
