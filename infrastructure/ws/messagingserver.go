package ws

type MessagingServer interface {
	Publish(channel string, data interface{}) error
	PublishToGroup(groupID uint, data interface{}) error
}
