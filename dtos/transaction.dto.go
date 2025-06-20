package sdk_dto

type (
	Transactions struct {
		ID                string  `json:"id"`
		TransactionID     string  `json:"transaction_id"`
		Amount            float64 `json:"amount"`
		Fee               float64 `json:"fee"`
		Bank              string  `json:"bank"`
		Channel           string  `json:"channel"`
		Status            string  `json:"status"`
		IsPaid            bool    `json:"is_paid"`
		RequestID         string  `json:"request_id"`
		ExternalID        string  `json:"external_id"`
		ExpiredAt         string  `json:"expired_at"`
		PaymentVerifiedAt string  `json:"payment_verified_at"`
		Notes             string  `json:"notes"`
		WebhookUrl        string  `json:"webhook_url"`
		AdditionalInfo    any     `json:"additional_info"`
		CreatedAt         string  `json:"created_at"`
		UpdatedAt         string  `json:"updated_at,omitempty"`
		PaymentLink       string  `json:"payment_link,omitempty"`
		QrImage           string  `json:"qr_image,omitempty"`
		VaNumber          string  `json:"va_number,omitempty"`
		NetworkID         string  `json:"network_id,omitempty"`
		PartnerID         string  `json:"partner_id,omitempty"`
		StatusMessage     string  `json:"status_message,omitempty"`
		NetworkReference  string  `json:"network_reference,omitempty"`
		NetworkData       any     `json:"network_data,omitempty"`
		AccountID         string  `json:"account_id,omitempty"`
		DepositID         string  `json:"deposit_id,omitempty"`
		FeeAmount         float64 `json:"fee_amount,omitempty"`
		TotalAmount       float64 `json:"total_amount,omitempty"`
		AddressNumber     string  `json:"address_number,omitempty"`
		AddressName       string  `json:"address_name,omitempty"`
		RequestData       any     `json:"request_data,omitempty"`
	}

	CallbackTransaction struct {
		TransactionID     string `json:"transaction_id"`
		PartnerID         string `json:"partner_id,omitempty"`
		Amount            any    `json:"amount"`
		FeeAmount         any    `json:"fee_amount"`
		Bank              string `json:"bank"`
		Channel           string `json:"channel"`
		ExternalID        string `json:"external_id"`
		Status            string `json:"status"`
		StatusMessage     string `json:"status_message"`
		NetworkData       any    `json:"network_data,omitempty"`
		NetworkReference  string `json:"network_reference,omitempty"`
		PaymentVerifiedAt string `json:"payment_verified_at"`
		IsPaid            bool   `json:"is_paid"`
		RequestData       any    `json:"request_data,omitempty"`
	}
)
