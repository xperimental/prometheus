// Copyright 2021 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package dockerswarm

import (
	"context"
	"fmt"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestDockerSDRefresh(t *testing.T) {
	sdmock := NewSDMock(t, "dockerprom")
	sdmock.Setup()

	e := sdmock.Endpoint()
	url := e[:len(e)-1]
	cfgString := fmt.Sprintf(`
---
host: %s
`, url)
	var cfg DockerSDConfig
	require.NoError(t, yaml.Unmarshal([]byte(cfgString), &cfg))

	d, err := NewDockerDiscovery(&cfg, log.NewNopLogger())
	require.NoError(t, err)

	ctx := context.Background()
	tgs, err := d.refresh(ctx)
	require.NoError(t, err)

	require.Equal(t, 1, len(tgs))

	tg := tgs[0]
	require.NotNil(t, tg)
	require.NotNil(t, tg.Targets)
	require.Equal(t, 1, len(tg.Targets))

	for i, lbls := range []model.LabelSet{
		{
			"__address__":                model.LabelValue("172.22.0.2:9100"),
			"__meta_docker_container_id": model.LabelValue("8bfd9b9a50425368797b3de45835fe5b8e21f479a2b90a847fdf265c0af9395a"),
			"__meta_docker_container_label_maintainer":                              model.LabelValue("The Prometheus Authors <prometheus-developers@googlegroups.com>"),
			"__meta_docker_container_label_prometheus_io_port":                      model.LabelValue("9100"),
			"__meta_docker_container_label_prometheus_io_scrape":                    model.LabelValue("yes"),
			"__meta_docker_container_name":                                          model.LabelValue("/dockersd_node_1"),
			"__meta_docker_container_network_mode":                                  model.LabelValue("dockersd_default"),
			"__meta_docker_network_ip":                                              model.LabelValue("172.22.0.2"),
			"__meta_docker_port_private":                                            model.LabelValue("9100"),
		},
	} {
		t.Run(fmt.Sprintf("item %d", i), func(t *testing.T) {
			require.Equal(t, lbls, tg.Targets[i])
		})
	}
}
