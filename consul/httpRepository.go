package consul

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

type httpRepo struct {
	url     string
	rawData []byte
	pool    *http.Client
}

//Create new http repository satisfying Repository interface
func NewHTTPRepository(url string) Repository {
	logrus.Infof("Creating http repository. URL: %s", url)
	return &httpRepo{url: url, pool: &http.Client{Timeout: time.Duration(5) * time.Second}}
}

//Load data from File and store in the memory
func (r *httpRepo) GetData() error {
	var resp *http.Response
	resp, err := r.pool.Get(r.url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return errors.New("response code is different than 200")
	}

	r.rawData, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return nil
}

//Parse loaded data and return as json.
func (r *httpRepo) ParseData() ([]map[string]interface{}, error) {
	logrus.Info("Repository: Parsing Data")
	var objmap []map[string]interface{}
	if err := json.Unmarshal(r.rawData, &objmap); err != nil {
		return nil, err
	}
	return objmap, nil
}
