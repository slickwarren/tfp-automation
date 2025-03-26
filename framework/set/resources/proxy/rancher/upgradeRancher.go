package rancher

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
	upgradeRancher = "upgrade_proxy_rancher"
)

// UpgradeProxiedRancher is a function that will upgrade the Rancher configurations in the main.tf file.
func UpgradeProxiedRancher(file *os.File, newFile *hclwrite.File, rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig,
	proxyPrivateIP, proxyNode string) (*os.File, error) {
	userDir := os.Getenv("GOROOT")

	scriptPath := filepath.Join(userDir, "src/github.com/rancher/tfp-automation/framework/set/resources/proxy/rancher/upgrade.sh")

	scriptContent, err := os.ReadFile(scriptPath)
	if err != nil {
		return nil, err
	}

	_, provisionerBlockBody := rke2.CreateNullResource(rootBody, terraformConfig, proxyNode, upgradeRancher)

	command := "bash -c '/tmp/upgrade.sh " + terraformConfig.Standalone.UpgradedRancherChartRepository + " " +
		terraformConfig.Standalone.UpgradedRancherRepo + " " + terraformConfig.Standalone.RancherHostname + " " + terraformConfig.Standalone.UpgradedRancherTagVersion + " " +
		terraformConfig.Standalone.UpgradedRancherImage + " " + proxyPrivateIP

	if terraformConfig.Standalone.UpgradedRancherAgentImage != "" {
		command += " " + terraformConfig.Standalone.UpgradedRancherAgentImage
	}

	command += " || true'"

	provisionerBlockBody.SetAttributeValue(defaults.Inline, cty.ListVal([]cty.Value{
		cty.StringVal("printf '" + string(scriptContent) + "' > /tmp/upgrade.sh"),
		cty.StringVal("chmod +x /tmp/upgrade.sh"),
		cty.StringVal(command),
	}))

	_, err = file.Write(newFile.Bytes())
	if err != nil {
		logrus.Infof("Failed to append configurations to main.tf file. Error: %v", err)
		return nil, err
	}

	return file, nil
}
