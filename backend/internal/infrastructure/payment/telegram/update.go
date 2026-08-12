package telegram

// Update is a subset of Telegram Bot API Update used for payments.
type Update struct {
	UpdateID          int64              `json:"update_id"`
	Message           *Message           `json:"message"`
	PreCheckoutQuery  *PreCheckoutQuery  `json:"pre_checkout_query"`
	CallbackQuery     *CallbackQuery     `json:"callback_query"`
}

// Message is a Telegram message.
type Message struct {
	MessageID         int64              `json:"message_id"`
	Text              string             `json:"text"`
	Chat              *Chat              `json:"chat"`
	From              *User              `json:"from"`
	SuccessfulPayment *SuccessfulPayment `json:"successful_payment"`
}

// Chat is a Telegram chat.
type Chat struct {
	ID int64 `json:"id"`
}

// User is a Telegram user.
type User struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
}

// PreCheckoutQuery is sent before Stars payment capture.
type PreCheckoutQuery struct {
	ID             string `json:"id"`
	From           *User  `json:"from"`
	Currency       string `json:"currency"`
	TotalAmount    int    `json:"total_amount"`
	InvoicePayload string `json:"invoice_payload"`
}

// SuccessfulPayment is delivered after Stars payment.
type SuccessfulPayment struct {
	Currency                string `json:"currency"`
	TotalAmount             int    `json:"total_amount"`
	InvoicePayload          string `json:"invoice_payload"`
	TelegramPaymentChargeID string `json:"telegram_payment_charge_id"`
	ProviderPaymentChargeID string `json:"provider_payment_charge_id"`
}

// CallbackQuery is an inline button press.
type CallbackQuery struct {
	ID   string   `json:"id"`
	From *User    `json:"from"`
	Data string   `json:"data"`
	Message *Message `json:"message"`
}
