package main

import (
	"collectd.org/api"
	"collectd.org/exec"
	"collectd.org/network"
	"context"
	"fmt"
	"math/rand"
	"net"
	"time"
)

func main() {

	client, _ := network.Dial(
		net.JoinHostPort("127.0.0.1", network.DefaultService),
		network.ClientOptions{
			SecurityLevel: network.None,
		})

	var flip bool
	for {
		i := -1.1 * float64(rand.Intn(5))
		if flip {
			i = i * -1
		}
		_ = client.Write(context.Background(), &api.ValueList{
			Identifier: api.Identifier{
				Host:   exec.Hostname(),
				Plugin: fmt.Sprint("seth_test_gauge_", 1),
				Type:   "gauge",
			},
			Time:     time.Now(),
			Interval: time.Minute,
			Values:   []api.Value{api.Gauge(i)},
		})

		_ = client.Write(context.Background(), &api.ValueList{
			Identifier: api.Identifier{
				Host:   exec.Hostname(),
				Plugin: fmt.Sprint("seth_test_counter_", 1),
				Type:   "counter",
			},
			Time:     time.Now(),
			Interval: time.Minute,
			Values:   []api.Value{api.Counter(1)},
		})
		if flip {
			_ = client.Flush()
		}
		flip = !flip
		time.Sleep(time.Millisecond * time.Duration(rand.Intn(100)))
	}
}
