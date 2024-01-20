package messagesender

type MessageSender interface {
	SendMessage(message string) error
}
