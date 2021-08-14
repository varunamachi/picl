package cfx

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"os"

	"github.com/sirupsen/logrus"
)

func LoadJsonFile(path string, out interface{}) error {
	reader, err := os.Open(path)
	if err != nil {
		logrus.WithError(err).WithField("path", path).
			Error("Failed to open JSON file")
		return FileErrf(err, "Failed to open JSON file at %s", path)
	}
	if err = LoadJson(reader, out); err != nil {
		logrus.WithError(err).WithField("path", path)
		return FileErrf(err, "Failed to load JSON data from file at %s", path)
	}
	return nil
}

func LoadJson(reader io.Reader, out interface{}) error {
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		const msg = "Failed to read from reader"
		logrus.WithError(err).Error(msg)
		return FileErrf(err, msg)
	}

	if err = json.Unmarshal(data, out); err != nil {
		const msg = "Failed to decode JSON data"
		logrus.WithError(err).Error(msg)
		return FileErrf(err, "Failed to decode JSON data")
	}
	return nil
}
