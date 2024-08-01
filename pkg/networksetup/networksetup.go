package networksetup

import (
	"os/exec"
	"strconv"
	"strings"
)

// networksetup -listallnetworkservices
func ListallNetworkServices() ([]string, error) {
	cmd := exec.Command("networksetup", "-listallnetworkservices")
	bytes, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	output := string(bytes)
	lines := strings.Split(output, "\n")

	networkServices := []string{}
	for _, line := range lines {
		if strings.HasPrefix(line, "An asterisk") {
			continue
		}

		if strings.TrimSpace(line) == "" {
			continue
		}
		networkServices = append(networkServices, strings.TrimSpace(line))
	}

	return networkServices, nil
}

// networksetup -setsocksfirewallproxy <networkservice> <domain> <port number> <authenticated> <username> <password>
func SetSocksFirewallProxy(networkService string, domain string, port int) error {
	cmd := exec.Command("networksetup", "-setsocksfirewallproxy", networkService, domain, strconv.Itoa(port))
	return cmd.Run()
}

// networksetup -setsocksfirewallproxystate <networkservice> <on off>
func SetSocksFirewallProxyState(networkService string, toggle bool) error {
	toggleStr := "off"
	if toggle {
		toggleStr = "on"
	}

	cmd := exec.Command("networksetup", "-setsocksfirewallproxystate", networkService, toggleStr)
	return cmd.Run()
}
