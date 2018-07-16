package runtime

func NewJSONBaseResourceVersioner() ResourceVersioner {
	return new(jsonBaseResourceVersioner)
}

type jsonBaseResourceVersioner struct{}

func (j *jsonBaseResourceVersioner) SetResourceVersion(obj Object, version uint64) error {
	// XXX
	return nil
}

func (j *jsonBaseResourceVersioner) ResourceVersion(obj Object) (uint64, error) {
	panic("not implemented")
}
