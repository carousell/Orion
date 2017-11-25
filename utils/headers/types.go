package headers

type Headers interface {
	Add(key, value string)
	Del(key string)
	Get(key string) string
	GetAll() map[string]string
}
