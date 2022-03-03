package consul

//Interface to provide different repository support
type Repository interface {
	GetData() error
	ParseData() ([]map[string]interface{}, error)
}

//Interface to provide higher abstraction layer
type Service interface {
	LoadData() error
	GetNodeInformationByHostname(hostname string) (string, error)
	ServiceAddressesByServiceTag(serviceTag string) ([]string, error)
	CheckRuleEligibility(serviceTag string, hostname string) bool
}
