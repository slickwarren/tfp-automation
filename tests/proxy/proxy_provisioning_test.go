package proxy

import (
	"os"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/rancher/tests/v2/actions/pipeline"
	"github.com/rancher/shepherd/clients/rancher"
	management "github.com/rancher/shepherd/clients/rancher/generated/management/v3"
	"github.com/rancher/shepherd/extensions/token"
	shepherdConfig "github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/shepherd/pkg/config/operations"
	"github.com/rancher/shepherd/pkg/session"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/configs"
	"github.com/rancher/tfp-automation/defaults/keypath"
	"github.com/rancher/tfp-automation/defaults/modules"
	"github.com/rancher/tfp-automation/framework"
	"github.com/rancher/tfp-automation/framework/cleanup"
	resources "github.com/rancher/tfp-automation/framework/set/resources/proxy"
	"github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	qase "github.com/rancher/tfp-automation/pipeline/qase/results"
	"github.com/rancher/tfp-automation/tests/extensions/provisioning"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type TfpProxyProvisioningTestSuite struct {
	suite.Suite
	client                     *rancher.Client
	session                    *session.Session
	cattleConfig               map[string]any
	rancherConfig              *rancher.Config
	terraformConfig            *config.TerraformConfig
	terratestConfig            *config.TerratestConfig
	standaloneTerraformOptions *terraform.Options
	terraformOptions           *terraform.Options
	proxyBastion               string
}

func (p *TfpProxyProvisioningTestSuite) TearDownSuite() {
	keyPath := rancher2.SetKeyPath(keypath.ProxyKeyPath)
	cleanup.Cleanup(p.T(), p.standaloneTerraformOptions, keyPath)
}

func (p *TfpProxyProvisioningTestSuite) SetupSuite() {
	p.cattleConfig = shepherdConfig.LoadConfigFromFile(os.Getenv(shepherdConfig.ConfigEnvironmentKey))
	p.rancherConfig, p.terraformConfig, p.terratestConfig = config.LoadTFPConfigs(p.cattleConfig)

	keyPath := rancher2.SetKeyPath(keypath.ProxyKeyPath)
	standaloneTerraformOptions := framework.Setup(p.T(), p.terraformConfig, p.terratestConfig, keyPath)
	p.standaloneTerraformOptions = standaloneTerraformOptions

	proxyBastion, _, err := resources.CreateMainTF(p.T(), p.standaloneTerraformOptions, keyPath, p.terraformConfig, p.terratestConfig)
	require.NoError(p.T(), err)

	p.proxyBastion = proxyBastion
}

func (p *TfpProxyProvisioningTestSuite) TfpSetupSuite() map[string]any {
	testSession := session.NewSession()
	p.session = testSession

	p.cattleConfig = shepherdConfig.LoadConfigFromFile(os.Getenv(shepherdConfig.ConfigEnvironmentKey))
	configMap, err := provisioning.UniquifyTerraform([]map[string]any{p.cattleConfig})
	require.NoError(p.T(), err)

	p.cattleConfig = configMap[0]
	p.rancherConfig, p.terraformConfig, p.terratestConfig = config.LoadTFPConfigs(p.cattleConfig)

	adminUser := &management.User{
		Username: "admin",
		Password: p.rancherConfig.AdminPassword,
	}

	userToken, err := token.GenerateUserToken(adminUser, p.rancherConfig.Host)
	require.NoError(p.T(), err)

	p.rancherConfig.AdminToken = userToken.Token

	client, err := rancher.NewClient(p.rancherConfig.AdminToken, testSession)
	require.NoError(p.T(), err)

	p.client = client
	p.client.RancherConfig.AdminToken = p.rancherConfig.AdminToken
	p.client.RancherConfig.AdminPassword = p.rancherConfig.AdminPassword
	p.client.RancherConfig.Host = p.rancherConfig.Host

	operations.ReplaceValue([]string{"rancher", "adminToken"}, p.rancherConfig.AdminToken, configMap[0])
	operations.ReplaceValue([]string{"rancher", "adminPassword"}, p.rancherConfig.AdminPassword, configMap[0])
	operations.ReplaceValue([]string{"rancher", "host"}, p.rancherConfig.Host, configMap[0])

	err = pipeline.PostRancherInstall(p.client, p.client.RancherConfig.AdminPassword)
	require.NoError(p.T(), err)

	keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath)
	terraformOptions := framework.Setup(p.T(), p.terraformConfig, p.terratestConfig, keyPath)
	p.terraformOptions = terraformOptions

	return p.cattleConfig
}

func (p *TfpProxyProvisioningTestSuite) TestTfpNoProxyProvisioning() {
	nodeRolesDedicated := []config.Nodepool{config.EtcdNodePool, config.ControlPlaneNodePool, config.WorkerNodePool}

	tests := []struct {
		name      string
		nodeRoles []config.Nodepool
		module    string
	}{
		{"No Proxy RKE1", nodeRolesDedicated, "ec2_rke1"},
		{"No Proxy RKE2", nodeRolesDedicated, "ec2_rke2"},
		{"No Proxy RKE2 Windows", nil, "ec2_rke2_windows_custom"},
		{"No Proxy K3S", nodeRolesDedicated, "ec2_k3s"},
	}

	for _, tt := range tests {
		cattleConfig := p.TfpSetupSuite()
		configMap := []map[string]any{cattleConfig}

		operations.ReplaceValue([]string{"terratest", "nodepools"}, tt.nodeRoles, configMap[0])
		operations.ReplaceValue([]string{"terraform", "module"}, tt.module, configMap[0])
		operations.ReplaceValue([]string{"terraform", "proxy", "proxyBastion"}, "", configMap[0])

		provisioning.GetK8sVersion(p.T(), p.client, p.terratestConfig, p.terraformConfig, configs.DefaultK8sVersion, configMap)

		_, terraform, terratest := config.LoadTFPConfigs(configMap[0])

		tt.name = tt.name + " Kubernetes version: " + terratest.KubernetesVersion
		testUser, testPassword := configs.CreateTestCredentials()

		p.Run((tt.name), func() {
			keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath)
			defer cleanup.Cleanup(p.T(), p.terraformOptions, keyPath)

			clusterIDs := provisioning.Provision(p.T(), p.client, p.rancherConfig, terraform, terratest, testUser, testPassword, p.terraformOptions, configMap, false)
			provisioning.VerifyClustersState(p.T(), p.client, clusterIDs)

			if strings.Contains(terraform.Module, modules.CustomEC2RKE2Windows) {
				clusterIDs := provisioning.Provision(p.T(), p.client, p.rancherConfig, terraform, terratest, testUser, testPassword, p.terraformOptions, configMap, true)
				provisioning.VerifyClustersState(p.T(), p.client, clusterIDs)
			}
		})
	}

	if p.terratestConfig.LocalQaseReporting {
		qase.ReportTest()
	}
}

func (p *TfpProxyProvisioningTestSuite) TestTfpProxyProvisioning() {
	nodeRolesDedicated := []config.Nodepool{config.EtcdNodePool, config.ControlPlaneNodePool, config.WorkerNodePool}

	tests := []struct {
		name      string
		nodeRoles []config.Nodepool
		module    string
	}{
		{"Proxy RKE1", nodeRolesDedicated, "ec2_rke1"},
		{"Proxy RKE2", nodeRolesDedicated, "ec2_rke2"},
		{"Proxy RKE2 Windows", nil, "ec2_rke2_windows_custom"},
		{"Proxy K3S", nodeRolesDedicated, "ec2_k3s"},
	}

	for _, tt := range tests {
		cattleConfig := p.TfpSetupSuite()
		configMap := []map[string]any{cattleConfig}

		operations.ReplaceValue([]string{"terratest", "nodepools"}, tt.nodeRoles, configMap[0])
		operations.ReplaceValue([]string{"terraform", "module"}, tt.module, configMap[0])
		operations.ReplaceValue([]string{"terraform", "proxy", "proxyBastion"}, p.proxyBastion, configMap[0])

		provisioning.GetK8sVersion(p.T(), p.client, p.terratestConfig, p.terraformConfig, configs.DefaultK8sVersion, configMap)

		_, terraform, terratest := config.LoadTFPConfigs(configMap[0])

		tt.name = tt.name + " Kubernetes version: " + terratest.KubernetesVersion
		testUser, testPassword := configs.CreateTestCredentials()

		p.Run((tt.name), func() {
			keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath)
			defer cleanup.Cleanup(p.T(), p.terraformOptions, keyPath)

			clusterIDs := provisioning.Provision(p.T(), p.client, p.rancherConfig, terraform, terratest, testUser, testPassword, p.terraformOptions, configMap, false)
			provisioning.VerifyClustersState(p.T(), p.client, clusterIDs)

			if strings.Contains(terraform.Module, modules.CustomEC2RKE2Windows) {
				clusterIDs := provisioning.Provision(p.T(), p.client, p.rancherConfig, terraform, terratest, testUser, testPassword, p.terraformOptions, configMap, true)
				provisioning.VerifyClustersState(p.T(), p.client, clusterIDs)
			}
		})
	}

	if p.terratestConfig.LocalQaseReporting {
		qase.ReportTest()
	}
}

func TestTfpProxyProvisioningTestSuite(t *testing.T) {
	suite.Run(t, new(TfpProxyProvisioningTestSuite))
}
