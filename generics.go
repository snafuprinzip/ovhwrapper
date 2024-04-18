package ovhwrapper

import (
	"encoding/json"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"path"
)

type StatusMsg interface {
	StatusMsg() string
}

func ToYaml[T any](object T) string {
	y, err := yaml.Marshal(&object)
	if err != nil {
		log.Printf("error marshalling yaml: %v", err)
	}
	return string(y)
}

func ToJSON[T any](object T) string {
	j, err := json.Marshal(&object)
	if err != nil {
		log.Printf("error marshalling json: %v", err)
	}
	return string(j)
}

func SaveYaml[T any](object T, fpath string) error {
	y, err := yaml.Marshal(&object)
	if err != nil {
		return err
	}

	// create directory if necessaru
	dir := path.Dir(fpath)
	err = os.MkdirAll(dir, 0700)
	if err != nil {
		log.Fatalf("Error creating %s directory: %v", dir, err)
	}

	err = os.WriteFile(fpath, y, 0644)
	if err != nil {
		return err
	}
	return nil
}

func LoadYaml[T any](object T, fpath string) error {
	srcFile, err := os.ReadFile(fpath)
	if err != nil {
		log.Printf("Can't read from %s: %v", fpath, err)
		return err
	}

	err = yaml.Unmarshal(srcFile, &object)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}
	return nil
}
