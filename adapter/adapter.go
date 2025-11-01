package adapter

type Adapter interface {
	Run() error
	Stop()
	SendResponse(response string)
}
