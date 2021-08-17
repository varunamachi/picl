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

// func DumpJSON(writer io.Writer, o interface{}) {
// 	b, err := json.MarshalIndent(o, "", "    ")
// 	if err == nil {
// 		fmt.Println(string(b))
// 	} else {
// 		LogErrorX("t.utils", "Failed to marshal data to JSON", err)
// 	}
// }

// //GetAsJSON - converts given data to JSON and returns as pretty printed
// func GetAsJSON(o interface{}) (jstr string, err error) {
// 	b, err := json.MarshalIndent(o, "", "    ")
// 	if err == nil {
// 		jstr = string(b)
// 	}
// 	return jstr, LogErrorX("t.utils", "Failed to marshal data to JSON", err)
// }

// //GetExecDir - gives absolute path of the directory in which the executable
// //for the current application is present
// func GetExecDir() (dirPath string) {
// 	execPath, err := os.Executable()
// 	if err == nil {
// 		dirPath = filepath.Dir(execPath)
// 	} else {
// 		LogErrorX("t.utils", "Failed to get the executable path", err)
// 	}

// 	return dirPath
// }

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
