package entity

type PushSubscription struct {
	Endpoint string
	Keys     PushKeys
}

type PushKeys struct {
	P256dh string
	Auth   string
}
