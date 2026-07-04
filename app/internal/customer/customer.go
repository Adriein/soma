package customer

type Customer struct {
	ID             int
	Name           string
	TelegramChatID int64
	Token          string
	TokenSecret    string
	TokenVerifier  int
}
