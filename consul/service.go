package consul

import (
	"errors"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"
)

type service struct {
	r       Repository
	content []map[string]interface{}
}

//Generates new service instance based on the repo
func NewService(repository Repository) Service {
	logrus.Info("Creating Consul Service")
	return &service{r: repository}
}

//Function calling underlying repository, retreive data and store as interface map
func (s *service) LoadData() error {
	logrus.Info("Loading Data")
	err := s.r.GetData()
	if err != nil {
		return err
	}

	s.content, err = s.r.ParseData()
	return err
}

//Function returning nodeParameters for specific hostname
func (s *service) GetNodeInformationByHostname(hostname string) (string, error) {
	logrus.Infof("Getting parameters for hostname: %s", hostname)
	for _, v := range s.content {
		if v["Node"] == hostname {
			return v["ServiceAddress"].(string), nil
		}
	}
	return "", errors.New("host not found")
}

//Check if rule is eligible to be applied on the host
func (s *service) CheckRuleEligibility(serviceTag string, serviceAddress string) bool {
	serviceAddresses, _ := s.ServiceAddressesByServiceTag(serviceTag)
	for _, serviceIp := range serviceAddresses {
		if serviceIp == serviceAddress {
			return true
		}
	}
	return false
}

//Get all service IPs by service tag or part of service tag.
func (s *service) ServiceAddressesByServiceTag(serviceTag string) ([]string, error) {
	//Normalise tags (lower case)
	serviceTag = strings.ToLower(serviceTag)
	regex, err := regexp.Compile(serviceTag)
	if err != nil {
		return nil, err
	}

	applicableHosts := make([]string, 0)
	for _, v := range s.content {
		//First case: ServiceTag == ALL, add IP to the list and continue
		if serviceTag == "all" {
			applicableHosts = append(applicableHosts, v["ServiceAddress"].(string))
			continue
		}
		//At this moment we should expect normal serviceTag (in dot format)
		serviceTags := v["ServiceTags"].([]interface{})

		for _, st := range serviceTags {

			if regex.MatchString(st.(string)) {
				applicableHosts = append(applicableHosts, v["ServiceAddress"].(string))
			}
		}
	}
	return applicableHosts, nil
}
