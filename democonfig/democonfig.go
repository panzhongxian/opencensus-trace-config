package democonfig

import (
	"fmt"
	"sync"
	"time"

	commonpb "github.com/census-instrumentation/opencensus-proto/gen-go/agent/common/v1"
	agenttracepb "github.com/census-instrumentation/opencensus-proto/gen-go/agent/trace/v1"
	tracepb "github.com/census-instrumentation/opencensus-proto/gen-go/trace/v1"
)

type DemoConfigUpdater struct{}

type NodeTraceConfig struct {
	config *agenttracepb.UpdatedLibraryConfig
	libver int64 // Library config version
}

var nodeTraceCfgMap = make(map[string]*NodeTraceConfig)
var nodeTraceVer int64
var startOnce sync.Once

func (cu *DemoConfigUpdater) Update(tcs agenttracepb.TraceService_ConfigServer, node *commonpb.Node) error {
	startOnce.Do(func() {
		go func() {
			for k, v := range nodeTraceCfgMap {
				if v != nil {
					continue
				}
				nodeTraceCfgMap[k] = &NodeTraceConfig{
					config: &agenttracepb.UpdatedLibraryConfig{
						Config: &tracepb.TraceConfig{
							Sampler: &tracepb.TraceConfig_ProbabilitySampler{
								ProbabilitySampler: &tracepb.ProbabilitySampler{
									SamplingProbability: 0.5,
								},
							},
						},
					},
					libver: nodeTraceVer,
				}
			}
			time.Sleep(1 * time.Millisecond)
		}()
	})

	for {
		// insert into the map when a new node comes.
		if value, ok := nodeTraceCfgMap[node.Identifier.HostName]; !ok {
			nodeTraceCfgMap[node.Identifier.HostName] = nil
			time.Sleep(1 * time.Millisecond)
			continue
		} else if value == nil {
			time.Sleep(1 * time.Millisecond)
			continue
		}

		// TODO: how to record version.
		err := tcs.Send(nodeTraceCfgMap[node.Identifier.HostName].config)
		if err != nil {
			fmt.Println("Send update config error: ", err)
			break
		}
		fmt.Println("After Send update config.")
		time.Sleep(1 * time.Second)
	}
	return nil
}
