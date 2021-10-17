package cfx

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"os"

	"github.com/sirupsen/logrus"
)

func LoadJsonFile(path string, out interface{}) error {
	reader, err := os.Open(path)
	if err != nil {
		logrus.WithError(err).WithField("path", path).
			Error("Failed to open JSON file")
		return Errf(err, "Failed to open JSON file at %s", path)
	}
	if err = LoadJson(reader, out); err != nil {
		logrus.WithError(err).WithField("path", path)
		return Errf(err, "Failed to load JSON data from file at %s", path)
	}
	return nil
}

func LoadJson(reader io.Reader, out interface{}) error {
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		const msg = "Failed to read from reader"
		logrus.WithError(err).Error(msg)
		return Errf(err, msg)
	}

	if err = json.Unmarshal(data, out); err != nil {
		const msg = "Failed to decode JSON data"
		logrus.WithError(err).Error(msg)
		return Errf(err, "Failed to decode JSON data")
	}
	return nil
}

//ExistsAsFile - checks if a regular file exists at given path. If a error
//occurs while stating whatever exists at given location, false is returned
func ExistsAsFile(path string) (yes bool) {
	stat, err := os.Stat(path)
	if err == nil && !stat.IsDir() {
		yes = true
	}
	return yes
}

//ExistsAsDir - checks if a directory exists at given path. If a error
//occurs while stating whatever exists at given location, false is returned
func ExistsAsDir(path string) (yes bool) {
	stat, err := os.Stat(path)
	if err == nil && stat.IsDir() {
		yes = true
	}
	return yes
}

func MustGetUserHome() string {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf(err.Error())
	}
	return home
}
