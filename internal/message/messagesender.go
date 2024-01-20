package message

type MessageSender interface {
	SendMessage(message string) error
}
