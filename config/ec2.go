package config

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

const ec2MetadataBase = "http://instance-data"

// GetEC2PrivIP queries EC2 metadata service to get the node's private IP
// address.
func (cfg *Cfg) GetEC2PrivIP() (ip string, err error) {
	if len(cfg.ec2PrivIPv4) == 0 {
		ep := fmt.Sprintf("%s/latest/meta-data/local-ipv4", cfg.ec2MetadataBase)

		var resp *http.Response
		resp, err = http.Get(ep)
		if err == nil {
			var ipb []byte
			ipb, err = ioutil.ReadAll(resp.Body)
			resp.Body.Close()

			if err == nil {
				cfg.ec2PrivIPv4 = string(ipb)
			}
		}
	}

	ip = cfg.ec2PrivIPv4

	return
}
