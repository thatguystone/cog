package node

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

// EC2Metadata provides wrapped and cached calls to the EC2 metadata service
type EC2Metadata struct {
	// Change this for testing to change which server is hit to get metadata
	base string

	privIP string
}

const ec2Base = "http://instance-data"

// EC2 provides a globally-accessible metadata instance
var EC2 EC2Metadata

func (e *EC2Metadata) addr(path string) string {
	b := ec2Base
	if e.base != "" {
		b = e.base
	}

	return fmt.Sprintf("%s%s", b, path)
}

// GetPrivIP gets the node's private IP address
func (e *EC2Metadata) GetPrivIP() (ip string, err error) {
	if e.privIP == "" {
		var resp *http.Response
		resp, err = http.Get(e.addr("/latest/meta-data/local-ipv4"))
		if err == nil {
			var ipb []byte
			ipb, err = ioutil.ReadAll(resp.Body)
			resp.Body.Close()

			if err == nil {
				e.privIP = string(ipb)
			}
		}
	}

	ip = e.privIP

	return
}
