package models

import (
    "time"
)

type Transaction struct {
    Signature  string    `json:"signature"`
    Timestamp  time.Time `json:"timestamp"`
    FromAddr   string    `json:"from_addr,omitempty"`
    ToAddr     string    `json:"to_addr,omitempty"`
    Amount     float64   `json:"amount,omitempty"`
    SlotNumber uint64    `json:"slot_number"`
}
