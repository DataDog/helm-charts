//go:build e2e

package datadog

import (
	"context"
	"fmt"
	"github.com/DataDog/helm-charts/test/common"
	"github.com/DataDog/test-infra-definitions/common/config"
	"github.com/DataDog/test-infra-definitions/resources/helm"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/stretchr/testify/require"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	gcpkubernetes "github.com/DataDog/datadog-agent/test/new-e2e/pkg/provisioners/gcp/kubernetes"
	kubeComp "github.com/DataDog/test-infra-definitions/components/kubernetes"

	"github.com/DataDog/datadog-agent/test/new-e2e/pkg/e2e"
	"github.com/DataDog/datadog-agent/test/new-e2e/pkg/environments"
)

type gkeSuite struct {
	e2e.BaseSuite[environments.Kubernetes]
}

func TestGKESuite(t *testing.T) {
	runnerConfig, err := common.SetupConfig()
	if err != nil {
		t.Skipf("Skipping test, problem setting up stack config: %s", err)
	}

	e2e.Run(t, &gkeSuite{}, e2e.WithProvisioner(gcpkubernetes.GKEProvisioner(gcpkubernetes.WithExtraConfigParams(runnerConfig), gcpkubernetes.WithWorkloadApp(datadogHelmInstallFunc))), e2e.WithSkipDeleteOnFailure(), e2e.WithDevMode())
}

func (v *gkeSuite) TestGKE() {
	v.T().Log("Running GKE test")
	res, _ := v.Env().KubernetesCluster.Client().CoreV1().Pods("datadog").List(context.TODO(), metav1.ListOptions{})

	var agent corev1.Pod
	containsAgent := false
	for _, pod := range res.Items {
		if strings.Contains(pod.Name, "dda-linux-datadog") && !strings.Contains(pod.Name, "cluster-agent") {
			containsAgent = true
			agent = pod
			break
		}
	}
	assert.True(v.T(), containsAgent, "Agent not found")

	stdout, stderr, err := v.Env().KubernetesCluster.KubernetesClient.
		PodExec("datadog", agent.Name, "agent", []string{"agent", "status"})
	require.NoError(v.T(), err)
	assert.Empty(v.T(), stderr)
	assert.NotEmpty(v.T(), stdout)

	var clusterAgent corev1.Pod
	containsClusterAgent := false
	for _, pod := range res.Items {
		if strings.Contains(pod.Name, "cluster-agent") {
			containsClusterAgent = true
			clusterAgent = pod
			break
		}
	}
	assert.True(v.T(), containsClusterAgent, "Cluster Agent not found")

	stdout, stderr, err = v.Env().KubernetesCluster.KubernetesClient.
		PodExec("datadog", clusterAgent.Name, "cluster-agent", []string{"agent", "status"})
	require.NoError(v.T(), err)
	assert.Empty(v.T(), stderr)
	assert.NotEmpty(v.T(), stdout)
}

func datadogHelmInstallFunc(e config.Env, kubeProvider *kubernetes.Provider) (*kubeComp.Workload, error) {
	var opts []pulumi.ResourceOption
	opts = append(opts, pulumi.Providers(kubeProvider), pulumi.DeletedWith(kubeProvider))

	rootDir := os.Getenv("CI_PROJECT_DIR")
	err := e.Ctx().Log.Info(fmt.Sprintf("WHAT IS ROOT DIR: %s", rootDir), nil)
	if err != nil {
		return nil, err
	}
	err = e.Ctx().Log.Info(fmt.Sprintf("WHAT IS CHART DIR: %s", path.Join(rootDir, "helm-charts", "charts", "datadog")), nil)
	if err != nil {
		return nil, err
	}
	_, err = helm.NewInstallation(e, helm.InstallArgs{
		RepoURL:     path.Join(rootDir, "helm-charts", "charts", "datadog"),
		ChartName:   "datadog",
		InstallName: "dda-linux-datadog",
		Namespace:   "datadog",
		ValuesYAML:  nil,
	}, opts...)
	if err != nil {
		return nil, err
	}
	return nil, nil
}
