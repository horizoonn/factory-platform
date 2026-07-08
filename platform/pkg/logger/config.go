package logger

const defaultLevel = "info"

type Config struct {
	Level       string
	JSON        bool
	Development bool
	ServiceName string
}
