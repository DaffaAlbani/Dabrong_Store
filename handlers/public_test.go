package handlers

import (
	"testing"
)

func TestParseAmountFromText(t *testing.T) {
	tests := []struct {
		input    string
		expected int
		wantErr  bool
	}{
		{"Berhasil menerima uang Rp 10.432 dari John Doe", 10432, false},
		{"Anda telah menerima transfer Rp. 10.432 dari...", 10432, false},
		{"Berhasil menerima uang Rp10.432 dari...", 10432, false},
		{"Transfer sebesar Rp. 10.432 berhasil masuk ke rekening...", 10432, false},
		{"DANA Bisnis: Pembayaran Rp 10.432 berhasil", 10432, false},
		{"Berhasil menerima uang Rp 1.500.000.", 1500000, false},
		{"Berhasil menerima uang Rp 500.", 500, false},
		{"Diterima Rp 50.000", 50000, false},
		{"Tidak ada nominal di sini", 0, true},
	}

	for _, tt := range tests {
		got, err := parseAmountFromText(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("parseAmountFromText(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			continue
		}
		if got != tt.expected {
			t.Errorf("parseAmountFromText(%q) = %v, want %v", tt.input, got, tt.expected)
		}
	}
}
