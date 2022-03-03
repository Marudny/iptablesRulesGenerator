package ipt

type Service interface {
	CreateNonexistingRuleset(chainName string, port string, ips []string) error
	CreateRuleset(port string, ips []string) error
	DeleteRuleset(chainName string, port string) error
	HasRulesetChanged(port string, ips []string) (bool, error)
	PurgeObsoleteChains(ports []string) error
	GetListOfConfiguredPorts() ([]string, error)
}
