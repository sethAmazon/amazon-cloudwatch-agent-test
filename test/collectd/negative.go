package main

import (
	"collectd.org/api"
	"collectd.org/exec"
	"collectd.org/network"
	"context"
	"fmt"
	"net"
	"time"
)

func main() {

	client, _ := network.Dial(
		net.JoinHostPort("127.0.0.1", network.DefaultService),
		network.ClientOptions{
			SecurityLevel: network.None,
		})

	for {
		_ = client.Write(context.Background(), &api.ValueList{
			Identifier: api.Identifier{
				Host:   exec.Hostname(),
				Plugin: fmt.Sprint("gauge_", 1),
				Type:   "gauge",
			},
			Time:     time.Now(),
			Interval: time.Minute,
			Values:   []api.Value{api.Gauge(-1.1)},
		})

		_ = client.Write(context.Background(), &api.ValueList{
			Identifier: api.Identifier{
				Host:   exec.Hostname(),
				Plugin: fmt.Sprint("counter_", 1),
				Type:   "counter",
			},
			Time:     time.Now(),
			Interval: time.Minute,
			Values:   []api.Value{api.Counter(-1)},
		})
		_ = client.Flush()
	}
}
