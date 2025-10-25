package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
)

type Item struct {
	Name       string `json:"name"`
	PriceCents int    `json:"price_cents"`
}

func main() {
	http.HandleFunc("/parse-receipt", func(w http.ResponseWriter, r *http.Request) {
		// For MVP we accept an image file via multipart form 'file' and return dummy items
		useLocal := os.Getenv("USE_LOCAL_OCR")
		_ = useLocal // future: switch between GCV and local

		items := []Item{
			{Name: "Milk", PriceCents: 12000},
			{Name: "Bread", PriceCents: 5000},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"items": items})
	})

	addr := ":8090"
	log.Printf("ocr-service listening on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}
