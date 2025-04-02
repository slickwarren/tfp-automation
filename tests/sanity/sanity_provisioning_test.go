package sanity

import (
	"os"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/shepherd/clients/rancher"
	management "github.com/rancher/shepherd/clients/rancher/generated/management/v3"
	"github.com/rancher/shepherd/extensions/token"
	shepherdConfig "github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/shepherd/pkg/config/operations"
	"github.com/rancher/shepherd/pkg/session"
	"github.com/rancher/tests/actions/pipeline"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/configs"
	"github.com/rancher/tfp-automation/defaults/keypath"
	"github.com/rancher/tfp-automation/defaults/modules"
	"github.com/rancher/tfp-automation/framework"
	"github.com/rancher/tfp-automation/framework/cleanup"
	"github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	resources "github.com/rancher/tfp-automation/framework/set/resources/sanity"
	qase "github.com/rancher/tfp-automation/pipeline/qase/results"
	"github.com/rancher/tfp-automation/tests/extensions/provisioning"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type TfpSanityProvisioningTestSuite struct {
	suite.Suite
	client                     *rancher.Client
	session                    *session.Session
	cattleConfig               map[string]any
	rancherConfig              *rancher.Config
	terraformConfig            *config.TerraformConfig
	terratestConfig            *config.TerratestConfig
	standaloneTerraformOptions *terraform.Options
	terraformOptions           *terraform.Options
}

func (s *TfpSanityProvisioningTestSuite) TearDownSuite() {
	keyPath := rancher2.SetKeyPath(keypath.SanityKeyPath, s.terraformConfig)
	cleanup.Cleanup(s.T(), s.standaloneTerraformOptions, keyPath)
}

func (s *TfpSanityProvisioningTestSuite) SetupSuite() {
	s.cattleConfig = shepherdConfig.LoadConfigFromFile(os.Getenv(shepherdConfig.ConfigEnvironmentKey))
	s.rancherConfig, s.terraformConfig, s.terratestConfig = config.LoadTFPConfigs(s.cattleConfig)

	keyPath := rancher2.SetKeyPath(keypath.SanityKeyPath, s.terraformConfig)
	standaloneTerraformOptions := framework.Setup(s.T(), s.terraformConfig, s.terratestConfig, keyPath)
	s.standaloneTerraformOptions = standaloneTerraformOptions

	_, err := resources.CreateMainTF(s.T(), s.standaloneTerraformOptions, keyPath, s.terraformConfig, s.terratestConfig)
	require.NoError(s.T(), err)
}

func (s *TfpSanityProvisioningTestSuite) TfpSetupSuite() map[string]any {
	testSession := session.NewSession()
	s.session = testSession

	s.cattleConfig = shepherdConfig.LoadConfigFromFile(os.Getenv(shepherdConfig.ConfigEnvironmentKey))
	configMap, err := provisioning.UniquifyTerraform([]map[string]any{s.cattleConfig})
	require.NoError(s.T(), err)

	s.cattleConfig = configMap[0]
	s.rancherConfig, s.terraformConfig, s.terratestConfig = config.LoadTFPConfigs(s.cattleConfig)

	adminUser := &management.User{
		Username: "admin",
		Password: s.rancherConfig.AdminPassword,
	}

	userToken, err := token.GenerateUserToken(adminUser, s.rancherConfig.Host)
	require.NoError(s.T(), err)

	s.rancherConfig.AdminToken = userToken.Token

	client, err := rancher.NewClient(s.rancherConfig.AdminToken, testSession)
	require.NoError(s.T(), err)

	s.client = client
	s.client.RancherConfig.AdminToken = s.rancherConfig.AdminToken
	s.client.RancherConfig.AdminPassword = s.rancherConfig.AdminPassword
	s.client.RancherConfig.Host = s.rancherConfig.Host

	operations.ReplaceValue([]string{"rancher", "adminToken"}, s.rancherConfig.AdminToken, configMap[0])
	operations.ReplaceValue([]string{"rancher", "adminPassword"}, s.rancherConfig.AdminPassword, configMap[0])
	operations.ReplaceValue([]string{"rancher", "host"}, s.rancherConfig.Host, configMap[0])

	err = pipeline.PostRancherInstall(s.client, s.client.RancherConfig.AdminPassword)
	require.NoError(s.T(), err)

	keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath, nil)
	terraformOptions := framework.Setup(s.T(), s.terraformConfig, s.terratestConfig, keyPath)
	s.terraformOptions = terraformOptions

	return s.cattleConfig
}

func (s *TfpSanityProvisioningTestSuite) TestTfpProvisioningSanity() {
	nodeRolesDedicated := []config.Nodepool{config.EtcdNodePool, config.ControlPlaneNodePool, config.WorkerNodePool}

	tests := []struct {
		name      string
		nodeRoles []config.Nodepool
		module    string
	}{
		{"Sanity RKE1", nodeRolesDedicated, "ec2_rke1"},
		{"Sanity RKE2", nodeRolesDedicated, "ec2_rke2"},
		{"Sanity RKE2 Windows", nil, "ec2_rke2_windows_custom"},
		{"Sanity K3S", nodeRolesDedicated, "ec2_k3s"},
	}

	for _, tt := range tests {
		cattleConfig := s.TfpSetupSuite()
		configMap := []map[string]any{cattleConfig}

		operations.ReplaceValue([]string{"terratest", "nodepools"}, tt.nodeRoles, configMap[0])
		operations.ReplaceValue([]string{"terraform", "module"}, tt.module, configMap[0])

		provisioning.GetK8sVersion(s.T(), s.client, s.terratestConfig, s.terraformConfig, configs.DefaultK8sVersion, configMap)

		_, terraform, terratest := config.LoadTFPConfigs(configMap[0])

		tt.name = tt.name + " Kubernetes version: " + terratest.KubernetesVersion
		testUser, testPassword := configs.CreateTestCredentials()

		s.Run((tt.name), func() {
			keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath, nil)
			defer cleanup.Cleanup(s.T(), s.terraformOptions, keyPath)

			clusterIDs := provisioning.Provision(s.T(), s.client, s.rancherConfig, terraform, terratest, testUser, testPassword, s.terraformOptions, configMap, false)
			provisioning.VerifyClustersState(s.T(), s.client, clusterIDs)

			if strings.Contains(terraform.Module, modules.CustomEC2RKE2Windows) {
				clusterIDs := provisioning.Provision(s.T(), s.client, s.rancherConfig, terraform, terratest, testUser, testPassword, s.terraformOptions, configMap, true)
				provisioning.VerifyClustersState(s.T(), s.client, clusterIDs)
			}
		})
	}

	if s.terratestConfig.LocalQaseReporting {
		qase.ReportTest()
	}
}

func TestTfpSanityProvisioningTestSuite(t *testing.T) {
	suite.Run(t, new(TfpSanityProvisioningTestSuite))
}
