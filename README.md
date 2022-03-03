## Short app description
App has been created in Golang (as I feel most comfortable with it). Apart of standard library modules, two external ones have been used:
- logrus (logger)
- coreos-iptables (iptables wrapper for GO)

### After start:
- gets hostname (it's used later to determine rules eligibility)
- initialises consul repository module (can be file or http based).
- loads data from json (from file or http response - depends on used repository)
 - gets IP address of local node from consul data (based on hostname)
 - initialize Firewall MetaRules (defined in task description). Rules are currently hardcoded, could be moved to a file or consul KV.
 - Initialise IPTables service
 - Go over rules, check if rule is applicable onto node, get all service IPs matching to the rule and create rulesets
 - Purge non-used chains. 

### IPTables management
#### Basic assumption:
- Work only on INPUT chain
- Create single chain per port in rulesets
- Chains have prefixes: auto_generated_port number
- In case of rule exchange first new chain with IPs is created and linked with INPUT chain, then old one is removed
- I truly believe that all possible scenarios have been covered and implemented.

#### Usage
When started without parameters it will try to connect to the consul service via http, get data and do the rest of work. If services.json is provided as argument application will read file and apply rules according to the file.
#### Build
To build application call:

    go build -mod vendor

#### Snippet from log
First run:
```
    sudo ./iptablesGenerator  services.json 
    INFO[0000] Firewall Creator                             
    INFO[0000] My hostname: node-01.eu-dc1.app.prod         
    INFO[0000] Creating file repository                     
    INFO[0000] Creating Consul Service                      
    INFO[0000] Loading Data                                 
    INFO[0000] Repository: Loading File services.json       
    INFO[0000] Repository: Parsing Data                     
    INFO[0000] Getting parameters for hostname: node-01.eu-dc1.app.prod 
    INFO[0000] My ip: 10.10.0.26                            
    INFO[0000] Creating Service                             
    INFO[0000] Ruleset for port 9100 doesn't exist          
    INFO[0000] Creating Rules for port 9100                 
    INFO[0000] Ruleset for port 9104 doesn't exist          
    INFO[0000] Creating Rules for port 9104                 
    INFO[0000] Ruleset for port 3306 doesn't exist          
    INFO[0000] Creating Rules for port 3306                 
    INFO[0000] Purging obsolete chains          
```
Another run:

```
    sudo ./iptablesGenerator services.json 
    INFO[0000] Firewall Creator                             
    INFO[0000] My hostname: node-01.eu-dc1.app.prod         
    INFO[0000] Creating repository                          
    INFO[0000] Creating Service                             
    INFO[0000] Service: Loading Data                        
    INFO[0000] Repository: Loading File services.json       
    INFO[0000] Repository: Parsing Data                     
    INFO[0000] Getting parameters for hostname: node-01.eu-dc1.app.prod 
    INFO[0000] My ip: 10.10.0.26                            
    INFO[0000] Creating Service                             
    INFO[0000] Ruleset for port 9100 hasn't been changed    
    INFO[0000] Ruleset for port 9104 hasn't been changed    
    INFO[0000] Ruleset for port 3306 hasn't been changed    
```
