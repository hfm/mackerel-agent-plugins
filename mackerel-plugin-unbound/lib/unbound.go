package mpunbound

import (
	"flag"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	mp "github.com/mackerelio/go-mackerel-plugin-helper"
	"github.com/mackerelio/golib/logging"
)

var logger = logging.GetLogger("metrics.plugin.unbound")

// UnboundPlugin mackerel plugin for Unbound
type UnboundPlugin struct {
	UnboundControlCommand string
	Ip                    string
	Port                  string
	Conf                  string
	Prefix                string
	EnableExtended        bool
	Tempfile              string
}

// MetricKeyPrefix interface for PluginWithPrefix
func (m UnboundPlugin) MetricKeyPrefix() string {
	if m.Prefix == "" {
		m.Prefix = "unbound"
	}
	return m.Prefix
}

// FetchMetrics interface for mackerelplugin
func (m UnboundPlugin) FetchMetrics() (map[string]interface{}, error) {
	server := fmt.Sprintf("%s@%s", m.Ip, m.Port)
	cmd := exec.Command(m.UnboundControlCommand, "-c", m.Conf, "-s", server)

	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		logger.Errorf("Failed to fetch statistics. %s", err)
		return nil, err
	}

	stat := make(map[string]interface{})

	keysStat := 0.0
	expiresStat := 0.0

	for _, line := range strings.Split(out.String(), "\r\n") {
		record := strings.Split(line, "=")
		if len(record) < 2 {
			continue
		}
		key, value := record[0], record[1]

		if re, _ := regexp.MatchString("^db", key); re {
			kv := strings.SplitN(value, ",", 3)
			keys, expires := kv[0], kv[1]

			keysKv := strings.SplitN(keys, "=", 2)
			keysFv, err := strconv.ParseFloat(keysKv[1], 64)
			if err != nil {
				logger.Warningf("Failed to parse db keys. %s", err)
			}
			keysStat += keysFv

			expiresKv := strings.SplitN(expires, "=", 2)
			expiresFv, err := strconv.ParseFloat(expiresKv[1], 64)
			if err != nil {
				logger.Warningf("Failed to parse db expires. %s", err)
			}
			expiresStat += expiresFv

			continue
		}

		stat[key], err = strconv.ParseFloat(value, 64)
		if err != nil {
			continue
		}
	}

	stat["keys"] = keysStat
	stat["expires"] = expiresStat

	if _, ok := stat["keys"]; !ok {
		stat["keys"] = 0
	}
	if _, ok := stat["expires"]; !ok {
		stat["expires"] = 0
	}

	if _, ok := stat["expired_keys"]; ok {
		stat["expired"] = stat["expired_keys"]
	} else {
		stat["expired"] = 0.0
	}

	return stat, nil
}

// GraphDefinition interface for mackerelplugin
func (m UnboundPlugin) GraphDefinition() map[string]mp.Graphs {
	labelPrefix := strings.Title(m.Prefix)

	var graphdef = map[string]mp.Graphs{
		"total": {
			Label: (labelPrefix + " Total"),
			Unit:  "integer",
			Metrics: []mp.Metrics{
				{Name: "total_commands_processed", Label: "Queries", Diff: true},
			},
		},
		"connections": {
			Label: (labelPrefix + " Connections"),
			Unit:  "integer",
			Metrics: []mp.Metrics{
				{Name: "total_connections_received", Label: "Connections", Diff: true, Stacked: true},
				{Name: "rejected_connections", Label: "Rejected Connections", Diff: true, Stacked: true},
			},
		},
		"clients": {
			Label: (labelPrefix + " Clients"),
			Unit:  "integer",
			Metrics: []mp.Metrics{
				{Name: "connected_clients", Label: "Connected Clients", Diff: false, Stacked: true},
				{Name: "blocked_clients", Label: "Blocked Clients", Diff: false, Stacked: true},
				{Name: "connected_slaves", Label: "Connected Slaves", Diff: false, Stacked: true},
			},
		},
		"keys": {
			Label: (labelPrefix + " Keys"),
			Unit:  "integer",
			Metrics: []mp.Metrics{
				{Name: "keys", Label: "Keys", Diff: false},
				{Name: "expires", Label: "Keys with expiration", Diff: false},
				{Name: "expired", Label: "Expired Keys", Diff: false},
			},
		},
		"keyspace": {
			Label: (labelPrefix + " Keyspace"),
			Unit:  "integer",
			Metrics: []mp.Metrics{
				{Name: "keyspace_hits", Label: "Keyspace Hits", Diff: true},
				{Name: "keyspace_misses", Label: "Keyspace Missed", Diff: true},
			},
		},
		"memory": {
			Label: (labelPrefix + " Memory"),
			Unit:  "integer",
			Metrics: []mp.Metrics{
				{Name: "used_memory", Label: "Used Memory", Diff: false},
				{Name: "used_memory_rss", Label: "Used Memory RSS", Diff: false},
				{Name: "used_memory_peak", Label: "Used Memory Peak", Diff: false},
				{Name: "used_memory_lua", Label: "Used Memory Lua engine", Diff: false},
			},
		},
		"capacity": {
			Label: (labelPrefix + " Capacity"),
			Unit:  "percentage",
			Metrics: []mp.Metrics{
				{Name: "percentage_of_memory", Label: "Percentage of memory", Diff: false},
				{Name: "percentage_of_clients", Label: "Percentage of clients", Diff: false},
			},
		},
	}

	return graphdef
}

// Do the plugin
func Do() {
	optUnboundControlCommand := flag.String("command", "unbound-control", "Command path to unbound-control")
	optIp := flag.String("host", "127.0.0.1", "IPAddress")
	optPort := flag.String("port", "8953", "Port")
	optConf := flag.String("conf", "/etc/unbound/unbound.conf", "Config file")
	optEnableExtended := flag.Bool("enable_extended", false, "Enable extended statistics")
	optMetricKeyPrefix := flag.String("metric-key-prefix", "unbound", "metric key prefix")
	optTempfile := flag.String("tempfile", "", "Temp file name")
	flag.Parse()

	unbound := UnboundPlugin{
		UnboundControlCommand: *optUnboundControlCommand,
		Ip:       *optIp,
		Port:     *optPort,
		Conf:     *optConf,
		Prefix:   *optMetricKeyPrefix,
		Extended: *optExtended,
	}
	helper := mp.NewMackerelPlugin(unbound)
	helper.Tempfile = *optTempfile

	helper.Run()
}
