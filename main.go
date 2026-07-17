package main

import (
	"log"
	"os"
	"ml-topup-v2/appsetup"
)

func main() {
	app := appsetup.Setup()

	// ============================================================
	//  LISTEN PORT
	// ============================================================
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Printf("[SERVER] Server Full Next.js berjalan di http://localhost:%s\n", port)
	log.Fatal(app.Listen(":" + port))
}
