package port

// Commander is implemented by notifiers that support interactive commands
// (e.g. Discord slash commands, Telegram bot commands).
// Stdout does not implement this interface.
type Commander interface {
	RegisterCommands(provider StatusProvider)
}
