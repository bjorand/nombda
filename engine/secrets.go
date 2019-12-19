package engine

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

func ReadSecretFile(filename string) (map[string]string, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	secrets := make(map[string]string)
	if err := yaml.UnmarshalStrict(data, &secrets); err != nil {
		return nil, fmt.Errorf("Unable to validate yaml file %s: %s", filename, err.Error())
	}
	return secrets, nil
}
