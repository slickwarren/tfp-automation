package provisioning

import (
	"os"

	ranchFrame "github.com/rancher/rancher/tests/framework/pkg/config"
	"github.com/rancher/tfp-automation/config"
	set "github.com/rancher/tfp-automation/framework/set"
	"github.com/sirupsen/logrus"
)

const (
	terratest                = "terratest"
	terraformFrameworkConfig = "terraform"
)

// BuildModule is a function that builds the Terraform module.
func BuildModule() error {
	clusterConfig := new(config.TerratestConfig)
	ranchFrame.LoadConfig(terratest, clusterConfig)

	keyPath := set.SetKeyPath()

	err := set.SetConfigTF(clusterConfig, "")
	if err != nil {
		return err
	}

	module, err := os.ReadFile(keyPath + "/main.tf")
	if err != nil {
		logrus.Errorf("Failed to read main.tf file contents. Error: %v", err)
		return err
	}

	logrus.Infof(string(module))

	return nil
}
