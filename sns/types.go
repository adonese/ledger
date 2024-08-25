package main

import "github.com/adonese/ledger"

type EscrowTransactionWrapper struct {
	SystemTransactionID string  `json:"transaction_id,omitempty"`
	FromAccount         string  `json:"from_account,omitempty"`
	ToAccount           string  `json:"to_account,omitempty"`
	Amount              float64 `json:"amount"`
	Comment             string  `json:"comment,omitempty"`
	TransactionDate     int64   `json:"time,omitempty"`
	Status              string  `json:"status,omitempty"`

	InitiatorUUID   string             `json:"uuid,omitempty"`
	Timestamp       string             `json:"timestamp,omitempty"`
	SignedUUID      string             `json:"signed_uuid,omitempty"`
	CashoutProvider string             `json:"cashout_provider,omitempty"`
	Beneficiary     ledger.Beneficiary `json:"beneficiary,omitempty"`
	ServiceProvider string             `json:"service_provider,omitempty"`
}

func NewEscrowTransactionWrapper(tx ledger.EscrowTransaction) EscrowTransactionWrapper {
	return EscrowTransactionWrapper{
		SystemTransactionID: tx.SystemTransactionID,
		FromAccount:         tx.FromAccount,
		ToAccount:           tx.ToAccount[:3] + "****" + tx.ToAccount[len(tx.ToAccount)-4:],
		Amount:              tx.Amount,
		Comment:             tx.Comment,
		TransactionDate:     tx.TransactionDate,
		Status:              tx.Status.String(), // Convert Status to string

		InitiatorUUID:   tx.InitiatorUUID,
		Timestamp:       tx.Timestamp,
		SignedUUID:      tx.SignedUUID,
		CashoutProvider: tx.CashoutProvider,
		Beneficiary:     tx.Beneficiary,

		ServiceProvider: tx.ServiceProvider,
	}
}
