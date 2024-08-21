package metrics

type SecondsMultiplier float64

const (
	LogUnitSeconds      = SecondsMultiplier(1.0)
	LogUnitMilliseconds = SecondsMultiplier(1.0 / 1000.0)
	LogUnitMicroseconds = SecondsMultiplier(1.0 / 1000000.0)
)

type Configuration struct {
	RequestTimeLogUnit SecondsMultiplier
}

var defaultConfig = &Configuration{}

func Configure(c *Configuration) {
	defaultConfig = c
}
