package ipt

import (
	"regexp"
	"strings"

	"github.com/coreos/go-iptables/iptables"
	"github.com/sirupsen/logrus"
)

type service struct {
	h *iptables.IPTables
}

//Generates new service instance
func NewService() Service {
	logrus.Info("Creating Service")
	i, err := iptables.New()
	if err != nil {
		return nil
	}
	return &service{h: i}
}

//Create new ruleset and remove potential existing one.
func (s *service) CreateRuleset(port string, ips []string) error {
	logrus.Infof("Creating Rules for port %s", port)
	chainName := strings.Join([]string{"auto", "generated", port}, "_")

	exists, err := s.h.ChainExists("filter", chainName)
	if err != nil {
		return err
	}
	if exists {
		//Chain exists - create temporary chain, link into INPUT, remove old one and rename temp.
		tempName := strings.Join([]string{"auto", "generated", port, "temp"}, "_")
		err := s.CreateNonexistingRuleset(tempName, port, ips)
		if err != nil {
			return err
		}
		err = s.DeleteRuleset(chainName, port)
		if err != nil {
			return err
		}
		err = s.h.RenameChain("filter", tempName, chainName)
		if err != nil {
			return err
		}

	} else {
		//Chain doesn't exist. Let's create a new one
		err := s.CreateNonexistingRuleset(chainName, port, ips)
		if err != nil {
			return err
		}
	}
	return nil
}

//Create new ruleset and link it to the INPUT only.
func (s *service) CreateNonexistingRuleset(chainName string, port string, ips []string) error {
	err := s.h.NewChain("filter", chainName)
	if err != nil {
		return err
	}
	for _, ip := range ips {
		err := s.h.Append("filter", chainName, "-s", ip, "-j", "ACCEPT")
		if err != nil {
			return err
		}
	}

	//Add drop in case of default accept policy
	err = s.h.AppendUnique("filter", chainName, "-j", "DROP")
	if err != nil {
		return err
	}
	//Link newly created chain to INPUT
	err = s.h.AppendUnique("filter", "INPUT", "-p", "tcp", "--dport", port, "-j", chainName)
	if err != nil {
		return err
	}
	return nil
}

//Completely remove ruleset / chain
func (s *service) DeleteRuleset(chainName string, port string) error {
	//Unlink from INPUT
	logrus.Infof("Unlinking chain %s for port %s", chainName, port)

	err := s.h.DeleteIfExists("filter", "INPUT", "-p", "tcp", "--dport", port, "-j", chainName)
	if err != nil {
		return err
	}

	//Clear and delete chain
	logrus.Infof("Deleting chain %s", chainName)
	err = s.h.ClearAndDeleteChain("filter", chainName)
	if err != nil {
		return err
	}
	return nil
}

//Check if list of IPs in existing ruleset has changed
func (s *service) HasRulesetChanged(port string, ips []string) (bool, error) {
	chainName := strings.Join([]string{"auto", "generated", port}, "_")

	//Chain doesn't exists - we need to create a new one, hence returning true
	exists, err := s.h.ChainExists("filter", chainName)
	if err == nil && !exists {
		logrus.Infof("Ruleset for port %s doesn't exist", port)
		return true, nil
	}

	//Get iptables structured stats - it returns parsed structure with source IPs.
	stats, err := s.h.StructuredStats("filter", chainName)
	if err != nil {
		return false, err
	}

	//Create map of currently configured IPs
	existingIPs := make(map[string]bool, 0)
	for _, stat := range stats {
		//Ignore last DROP rule
		if stat.Source.IP.String() == "0.0.0.0" && stat.Target == "DROP" {
			continue
		}
		existingIPs[stat.Source.IP.String()] = false
	}

	//Find differences between lists.
	//First, compare lengths of both list - any difference means change.
	if len(existingIPs) != len(ips) {
		logrus.Infof("Ruleset for port %s has been changed", port)
		return true, nil
	}

	//Second, iterate over potential new list of IPs and compare with existing ones.
	for _, ip := range ips {
		//If key doesn't exist - IP is not already configured.
		if _, ok := existingIPs[ip]; !ok {
			logrus.Infof("Ruleset for port %s has been changed", port)
			return true, nil
		}
	}
	logrus.Infof("Ruleset for port %s hasn't been changed", port)
	return false, nil
}

//Get list of currently configured ports / chains
func (s *service) GetListOfConfiguredPorts() ([]string, error) {
	//Get list of existing chains.
	chains, err := s.h.ListChains("filter")
	if err != nil {
		return nil, err
	}
	//Regexp to filter out values
	regex := regexp.MustCompile(`auto_generated_(?P<Port>.*)`)

	//Create a list of configured ports
	configuredPorts := make([]string, 0)

	for _, chain := range chains {
		r := regex.FindStringSubmatch(chain)
		if r == nil {
			continue
		}
		portIndex := regex.SubexpIndex("Port")
		configuredPorts = append(configuredPorts, r[portIndex])
	}
	return configuredPorts, nil
}

//Find and remove obsolete chains
func (s *service) PurgeObsoleteChains(ports []string) error {
	logrus.Info("Purging obsolete chains")
	//Get list of existing chains.
	configuredPorts, err := s.GetListOfConfiguredPorts()
	if err != nil {
		return err
	}

	//Create a bool map.
	configuredPortsMap := make(map[string]bool, 0)

	//Go over list and add ports to the map
	for _, port := range configuredPorts {
		configuredPortsMap[port] = false
	}

	//Go over expected port list and mark all existing chains
	for _, port := range ports {
		//If key doesn't exist - port is not already configured.
		if _, ok := configuredPortsMap[port]; ok {
			configuredPortsMap[port] = true
		}
	}

	//Remove all not marked chains
	for port, v := range configuredPortsMap {
		if !v {
			chainName := strings.Join([]string{"auto", "generated", port}, "_")
			s.DeleteRuleset(chainName, port)
		}
	}

	return nil
}
