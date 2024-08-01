package v2ray

type Inbound struct {
	Protocol string `json:"protocol"`
	Port     int    `json:"port"`
	Listen   string `json:"listen"`
}

type Configuration struct {
	Inbounds []Inbound `json:"inbounds"`
}

func (conf *Configuration) FindSocksProxy() *Inbound {
	for idx := range conf.Inbounds {
		inbound := conf.Inbounds[idx]
		if inbound.Protocol == "socks" {
			return &inbound
		}
	}
	return nil
}
