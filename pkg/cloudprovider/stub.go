package cloudprovider

import (
	"errors"
	"io"
)

type Interface interface{}

func GetCloudProvider(name string, config io.Reader) (Interface, error) {
	return nil, errors.New("not implement yet")
}
