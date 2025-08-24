package headers

type Headers map[string]string

func NewHeaders() Headers {
	return make(Headers)
}
