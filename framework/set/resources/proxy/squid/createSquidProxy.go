package squid

import (
	"os"
	"path/filepath"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	"github.com/rancher/tfp-automation/framework/set/resources/rke2"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

const (
	installSquidProxy = "install_squid_proxy"
)

// CreateSquidProxy is a function that will set the squid proxy configurations in the main.tf file.
func CreateSquidProxy(file *os.File, newFile *hclwrite.File, rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig,
	rke2BastionPublicDNS, rke2ServerOnePrivateIP, rke2ServerTwoPrivateIP, rke2ServerThreePrivateIP string) (*os.File, error) {

	userDir := os.Getenv("GOROOT")

	scriptPath := filepath.Join(userDir, "src/github.com/rancher/tfp-automation/framework/set/resources/proxy/squid/setup.sh")
	squidConf := filepath.Join(userDir, "src/github.com/rancher/tfp-automation/framework/set/resources/proxy/squid/squid.conf")

	scriptContent, err := os.ReadFile(scriptPath)
	if err != nil {
		return nil, err
	}

	privateKey, err := os.ReadFile(terraformConfig.PrivateKeyPath)
	if err != nil {
		return nil, err
	}

	squidConfContent, err := os.ReadFile(squidConf)
	if err != nil {
		return nil, err
	}

	_, provisionerBlockBody := rke2.CreateNullResource(rootBody, terraformConfig, rke2BastionPublicDNS, installSquidProxy)

	command := "bash -c '/tmp/setup.sh " + terraformConfig.Standalone.OSUser + " " + terraformConfig.Standalone.OSGroup + " " +
		rke2BastionPublicDNS + " " + terraformConfig.Standalone.BootstrapPassword + " " + terraformConfig.StandaloneRegistry.RegistryUsername + " " +
		terraformConfig.StandaloneRegistry.RegistryPassword + " " + terraformConfig.StandaloneRegistry.RegistryName + " " +
		terraformConfig.Standalone.RKE2Version + " " + rke2ServerOnePrivateIP + " " + rke2ServerTwoPrivateIP + " " +
		rke2ServerThreePrivateIP + " || true'"

	provisionerBlockBody.SetAttributeValue(defaults.Inline, cty.ListVal([]cty.Value{
		cty.StringVal("echo '" + string(scriptContent) + "' > /tmp/setup.sh"),
		cty.StringVal("echo '" + string(squidConfContent) + "' > /tmp/squid.conf"),
		cty.StringVal("echo '" + string(privateKey) + "' > /tmp/keyfile.pem"),
		cty.StringVal("chmod +x /tmp/setup.sh"),
		cty.StringVal(command),
	}))

	_, err = file.Write(newFile.Bytes())
	if err != nil {
		logrus.Infof("Failed to append configurations to main.tf file. Error: %v", err)
		return nil, err
	}

	return file, nil
}
