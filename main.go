package main

import (
	"os"

	"github.com/sirupsen/logrus"
	"status.im/iptablesGenerator/v2/consul"
	"status.im/iptablesGenerator/v2/ipt"
)

func main() {
	logrus.Info("Firewall Creator")

	hostname, err := os.Hostname()
	if err != nil {
		logrus.Fatal("Ooopsie, cannot determine hostname. Exiting.")
	}
	//hostname = "node-01.eu-dc1.app.prod"
	logrus.Infof("My hostname: %s", hostname)

	//Create consul repository variable and read data from file
	var cRepo consul.Repository
	//If json file name is provided as a argument - read from file, otherwise create http repository
	if len(os.Args) > 1 {
		filename := os.Args[1]
		cRepo = consul.NewFileRepository(filename)
	} else {
		cRepo = consul.NewHTTPRepository("http://localhost:8500/v1/catalog/service/wireguard")
	}

	//Create consul service and parse data from Repository
	cService := consul.NewService(cRepo)
	err = cService.LoadData()
	if err != nil {
		logrus.Fatal(err)
	}
	//Get local node address from consul based on hostname
	localServiceIp, _ := cService.GetNodeInformationByHostname(hostname)
	logrus.Infof("My ip: %s\n", localServiceIp)

	//Create ruleset slice.
	rules := make(ipt.Ruleset, 0)
	rules = append(rules, ipt.Rule{From: "ALL", To: "logs.*", Port: "5141"})
	rules = append(rules, ipt.Rule{From: "metrics.*", To: "ALL", Port: "9100"})
	rules = append(rules, ipt.Rule{From: "metrics.*", To: "app.*", Port: "9104"})
	rules = append(rules, ipt.Rule{From: "backups.*", To: "app.*", Port: "3306"})

	//Create IPTables Service
	iptService := ipt.NewService()

	//Slice to keep list of applied ports
	var appliedPorts []string

	//Go over rule list and check if rule is applicable to the host
	for _, rule := range rules {
		//Check if rule is applicable on local node.
		if cService.CheckRuleEligibility(rule.To, localServiceIp) {
			appliedPorts = append(appliedPorts, rule.Port)
			//Get all serviceIPs based on from field.
			serviceIps, err := cService.ServiceAddressesByServiceTag(rule.From)
			if err != nil {
				logrus.Fatalf("ERR: %s", err.Error())
			}
			//Check if chain exists or ruleset has changed.
			ruleChanged, err := iptService.HasRulesetChanged(rule.Port, serviceIps)
			if err != nil {
				logrus.Fatalf("ERR: %s", err.Error())
			}
			if ruleChanged {
				//Create new or rebuild existing chain with current
				err = iptService.CreateRuleset(rule.Port, serviceIps)
				if err != nil {
					logrus.Fatalf("ERR: %s", err.Error())
				}

			}
		}
	}
	//Get list of applied ports and remove obsolete chains
	iptService.PurgeObsoleteChains(appliedPorts)

}
