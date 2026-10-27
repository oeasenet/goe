package metrics

import (
	"runtime"
	"runtime/debug"

	"github.com/prometheus/client_golang/prometheus"
)

const goeModule = "go.oease.dev/goe/v2"

// newBuildInfo returns the <namespace>_build_info gauge: always 1, labelled
// with the GOE version, the Go version and the application's (APP_VERSION).
func newBuildInfo(namespace, appVersion string) prometheus.Collector {
	info := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "build_info",
		Help:      "Build information: the GOE, Go and application versions. Always 1.",
	}, []string{"goe_version", "go_version", "version"})
	info.WithLabelValues(goeVersion(), runtime.Version(), appVersion).Set(1)
	return info
}

// goeVersion reports the GOE module version the binary was built with.
func goeVersion() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "unknown"
	}
	if info.Main.Path == goeModule {
		return info.Main.Version
	}
	for _, dep := range info.Deps {
		if dep.Path == goeModule {
			if dep.Replace != nil && dep.Replace.Version != "" {
				return dep.Replace.Version
			}
			return dep.Version
		}
	}
	return "unknown"
}
