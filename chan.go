package cog

// Notify sends a struct{} down the channel without waiting
func Notify(ch chan<- struct{}) {
	select {
	case ch <- struct{}{}:
	default:
	}
}
