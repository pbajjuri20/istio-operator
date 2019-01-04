package istiocontrolplane

import (
	"k8s.io/helm/pkg/proto/hapi/release"
	"k8s.io/helm/pkg/tiller"
	"fmt"
	"strings"

	"github.com/ghodss/yaml"

	"github.com/maistra/istio-operator/pkg/apis/istio/v1alpha3"

	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/manifest"
	"k8s.io/helm/pkg/proto/hapi/chart"
	"k8s.io/helm/pkg/renderutil"
	"k8s.io/helm/pkg/timeconv"
)

var (
	// ChartPath to helm charts
	ChartPath string
)

// RenderHelmChart renders the helm charts, returning a map of rendered templates.
// key names represent the chart from which the template was processed.  Subcharts
// will be keyed as <root-name>/charts/<subchart-name>, e.g. istio/charts/galley.
// The root chart would be simply, istio.
func RenderHelmChart(chartPath string, icp *v1alpha3.IstioControlPlane) (map[string][]manifest.Manifest, *release.Release, error) {
	rawVals, err := yaml.Marshal(icp.Spec)
	config := &chart.Config{Raw: string(rawVals), Values: map[string]*chart.Value{}}

	c, err := chartutil.Load(chartPath)
	if err != nil {
		return map[string][]manifest.Manifest{}, nil, err
	}

	renderOpts := renderutil.Options{
		ReleaseOptions: chartutil.ReleaseOptions{
			// XXX: hard code or use icp.GetName()
			Name:      "istio",
			IsInstall: true,
			IsUpgrade: false,
			Time:      timeconv.Now(),
			Namespace: icp.GetNamespace(),
		},
		// XXX: hard-code or look this up somehow?
		KubeVersion: fmt.Sprintf("%s.%s", chartutil.DefaultKubeVersion.Major, chartutil.DefaultKubeVersion.Minor),
	}
	renderedTemplates, err := renderutil.Render(c, config, renderOpts)
	if err != nil {
		return map[string][]manifest.Manifest{}, nil, err
	}

	rel := &release.Release{
		Name:      renderOpts.ReleaseOptions.Name,
		Chart:     c,
		Config:    config,
		Namespace: icp.GetNamespace(),
		Info:      &release.Info{LastDeployed: renderOpts.ReleaseOptions.Time},
	}

	return sortManifestsByChart(manifest.SplitManifests(renderedTemplates)), rel, nil
}

// sortManifestsByChart returns a map of chart->[]manifest.  names for subcharts
// will be of the form <root-name>/charts/<subchart-name>, e.g. istio/charts/galley
func sortManifestsByChart(manifests []manifest.Manifest) map[string][]manifest.Manifest {
	manifestsByChart := make(map[string][]manifest.Manifest)
	for _, chartManifest := range manifests {
		pathSegments := strings.Split(chartManifest.Name, "/")
		chartName := pathSegments[0]
		// paths always start with the root chart name and always have a template
		// name, so we should be safe not to check length
		if pathSegments[1] == "charts" {
			// subcharts will have names like <root-name>/charts/<subchart-name>/...
			chartName = strings.Join(pathSegments[:3], "/")
		}
		chartManifests, ok := manifestsByChart[chartName]
		if !ok {
			chartManifests = make([]manifest.Manifest, 0, 10)
			manifestsByChart[chartName] = chartManifests
		}
		chartManifests = append(chartManifests, chartManifest)
	}
	for key, value := range manifestsByChart {
		manifestsByChart[key] = tiller.SortByKind(value)
	}
	return manifestsByChart
}
