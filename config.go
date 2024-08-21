package metrics

type SecondsMultiplier float64

const (
	LogUnitSeconds      = SecondsMultiplier(1.0)
	LogUnitMilliseconds = SecondsMultiplier(1.0 / 1000.0)
	LogUnitMicroseconds = SecondsMultiplier(1.0 / 1000000.0)
)

type LogFieldValueFunc func(map[string]interface{}) interface{}

func FirstLevelLogField(fieldName string) LogFieldValueFunc {
	return func(log map[string]interface{}) interface{} {
		return log[fieldName]
	}
}

func SecondLevelLogField(firstLevelFieldName, secondLevelFieldName string) LogFieldValueFunc {
	return func(log map[string]interface{}) interface{} {
		first, ok := log[firstLevelFieldName]
		if !ok {
			return nil
		}
		m, ok := first.(map[string]interface{})
		if !ok {
			return nil
		}
		return m[secondLevelFieldName]
	}
}

type Configuration struct {
	RequestTimeLogField LogFieldValueFunc
	RequestTimeLogUnit  SecondsMultiplier
}

var defaultConfig = &Configuration{}

func Configure(c *Configuration) {
	defaultConfig = c
}
