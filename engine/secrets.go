package engine

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

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

func ReadSecretFromEnv() map[string]string {
	secrets := make(map[string]string)
	for _, env := range os.Environ() {
		s := strings.Split(env, "=")
		k := s[0]
		v := strings.TrimPrefix(env, fmt.Sprintf("%s=", k))
		if strings.HasPrefix(k, "NOMBDA_SECRET_") {
			secrets[strings.ToLower(strings.TrimPrefix(k, "NOMBDA_SECRET_"))] = v
		}
	}
	return secrets
}
