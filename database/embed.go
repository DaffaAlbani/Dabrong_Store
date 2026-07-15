package database

import _ "embed"

//go:embed products.json
var EmbeddedProducts []byte
