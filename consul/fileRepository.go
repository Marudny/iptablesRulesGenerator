package consul

import (
	"encoding/json"
	"io/ioutil"

	"github.com/sirupsen/logrus"
)

type frepo struct {
	filename string
	rawData  []byte
}

//Create new file repository satisfying Repository interface
func NewFileRepository(filename string) Repository {
	logrus.Info("Creating file repository")
	return &frepo{filename: filename}
}

//Load data from File and store in the memory
func (r *frepo) GetData() error {
	var err error
	logrus.Infof("Repository: Loading File %s", r.filename)
	r.rawData, err = ioutil.ReadFile(r.filename)
	if err != nil {
		return err
	}

	return err
}

//Parse loaded data and return as json.
func (r *frepo) ParseData() ([]map[string]interface{}, error) {
	logrus.Info("Repository: Parsing Data")
	var objmap []map[string]interface{}
	if err := json.Unmarshal(r.rawData, &objmap); err != nil {
		return nil, err
	}
	return objmap, nil
}
