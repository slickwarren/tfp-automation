package harvester

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/format"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	"github.com/zclconf/go-cty/cty"
)

// CreateHarvesterInstances is a function that will set the Harvester instances configurations in the main.tf file.
func CreateHarvesterInstances(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, terratestConfig *config.TerratestConfig,
	hostnamePrefix string) {

	configBlockSecret := rootBody.AppendNewBlock(defaults.Resource, []string{defaults.KubernetesSecret, hostnamePrefix + "secret"})
	configBlockSecretBody := configBlockSecret.Body()

	configBlockSecretBody.SetAttributeValue(defaults.Name, cty.StringVal(hostnamePrefix+"secret"))
	configBlockSecretBody.SetAttributeValue(defaults.Namespace, cty.StringVal("default"))

	configBlock := rootBody.AppendNewBlock(defaults.Resource, []string{defaults.HarvesterVirtualMachine, hostnamePrefix})
	configBlockBody := configBlock.Body()

	// not sure what this does yet, but it was for aws
	if strings.Contains(terraformConfig.Module, "custom") {
		configBlockBody.SetAttributeValue(defaults.Count, cty.NumberIntVal(terratestConfig.NodeCount))
	}

	configBlockBody.SetAttributeValue(defaults.Ami, cty.StringVal(terraformConfig.HarvesterConfig.AMI))
	configBlockBody.SetAttributeValue(defaults.InstanceType, cty.StringVal(terraformConfig.HarvesterConfig.HarvesterInstanceType))
	configBlockBody.SetAttributeValue(defaults.SubnetId, cty.StringVal(terraformConfig.HarvesterConfig.HarvesterSubnetID))

	securityGroups := format.ListOfStrings(terraformConfig.HarvesterConfig.HarvesterSecurityGroups)
	configBlockBody.SetAttributeRaw(defaults.VpcSecurityGroupIds, securityGroups)
	configBlockBody.SetAttributeValue(defaults.KeyName, cty.StringVal(terraformConfig.HarvesterConfig.HarvesterKeyName))

	configBlockBody.AppendNewline()

	rootBlockDevice := configBlockBody.AppendNewBlock(defaults.RootBlockDevice, nil)
	rootBlockDeviceBody := rootBlockDevice.Body()

	if strings.Contains(hostnamePrefix, "registry") {
		rootBlockDeviceBody.SetAttributeValue(defaults.VolumeSize, cty.NumberIntVal(terraformConfig.HarvesterConfig.RegistryRootSize))
	} else {
		rootBlockDeviceBody.SetAttributeValue(defaults.VolumeSize, cty.NumberIntVal(terraformConfig.HarvesterConfig.HarvesterRootSize))
	}

	configBlockBody.AppendNewline()

	tagsBlock := configBlockBody.AppendNewBlock(defaults.Tags+" =", nil)
	tagsBlockBody := tagsBlock.Body()

	if strings.Contains(terraformConfig.Module, "custom") {
		expression := fmt.Sprintf(`"%s-${`+defaults.Count+`.`+defaults.Index+`}"`, terraformConfig.ResourcePrefix+"-"+hostnamePrefix)
		tags := hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte(expression)},
		}

		tagsBlockBody.SetAttributeRaw(defaults.Name, tags)
	} else {
		expression := fmt.Sprintf(`"%s`, terraformConfig.ResourcePrefix+"-"+hostnamePrefix+`"`)
		tags := hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte(expression)},
		}

		tagsBlockBody.SetAttributeRaw(defaults.Name, tags)
	}

	configBlockBody.AppendNewline()

	connectionBlock := configBlockBody.AppendNewBlock(defaults.Connection, nil)
	connectionBlockBody := connectionBlock.Body()

	connectionBlockBody.SetAttributeValue(defaults.Type, cty.StringVal(defaults.Ssh))
	connectionBlockBody.SetAttributeValue(defaults.User, cty.StringVal(terraformConfig.HarvesterConfig.HarvesterUser))

	hostExpression := defaults.Self + "." + defaults.PublicIp
	host := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(hostExpression)},
	}

	connectionBlockBody.SetAttributeRaw(defaults.Host, host)

	keyPathExpression := defaults.File + `("` + terraformConfig.PrivateKeyPath + `")`
	keyPath := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(keyPathExpression)},
	}

	connectionBlockBody.SetAttributeRaw(defaults.PrivateKey, keyPath)
	connectionBlockBody.SetAttributeValue(defaults.Timeout, cty.StringVal(terraformConfig.HarvesterConfig.Timeout))

	configBlockBody.AppendNewline()

	provisionerBlock := configBlockBody.AppendNewBlock(defaults.Provisioner, []string{defaults.RemoteExec})
	provisionerBlockBody := provisionerBlock.Body()

	provisionerBlockBody.SetAttributeValue(defaults.Inline, cty.ListVal([]cty.Value{
		cty.StringVal("echo Connected!!!"),
	}))
}

// CreateAirgappedHarvesterInstances is a function that will set the Harvester instances configurations in the main.tf file.
func CreateAirgappedHarvesterInstances(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, hostnamePrefix string) {
	configBlock := rootBody.AppendNewBlock(defaults.Resource, []string{defaults.AwsInstance, hostnamePrefix})
	configBlockBody := configBlock.Body()

	configBlockBody.SetAttributeValue(defaults.AssociatePublicIPAddress, cty.BoolVal(false))
	configBlockBody.SetAttributeValue(defaults.Ami, cty.StringVal(terraformConfig.HarvesterConfig.AMI))
	configBlockBody.SetAttributeValue(defaults.InstanceType, cty.StringVal(terraformConfig.HarvesterConfig.HarvesterInstanceType))
	configBlockBody.SetAttributeValue(defaults.SubnetId, cty.StringVal(terraformConfig.HarvesterConfig.HarvesterSubnetID))

	securityGroups := format.ListOfStrings(terraformConfig.HarvesterConfig.HarvesterSecurityGroups)
	configBlockBody.SetAttributeRaw(defaults.VpcSecurityGroupIds, securityGroups)
	configBlockBody.SetAttributeValue(defaults.KeyName, cty.StringVal(terraformConfig.HarvesterConfig.HarvesterKeyName))

	configBlockBody.AppendNewline()

	rootBlockDevice := configBlockBody.AppendNewBlock(defaults.RootBlockDevice, nil)
	rootBlockDeviceBody := rootBlockDevice.Body()
	rootBlockDeviceBody.SetAttributeValue(defaults.VolumeSize, cty.NumberIntVal(terraformConfig.HarvesterConfig.HarvesterRootSize))

	configBlockBody.AppendNewline()

	tagsBlock := configBlockBody.AppendNewBlock(defaults.Tags+" =", nil)
	tagsBlockBody := tagsBlock.Body()

	expression := fmt.Sprintf(`"%s`, terraformConfig.ResourcePrefix+"-"+hostnamePrefix+`"`)
	tags := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(expression)},
	}

	tagsBlockBody.SetAttributeRaw(defaults.Name, tags)

	configBlockBody.AppendNewline()

	connectionBlock := configBlockBody.AppendNewBlock(defaults.Connection, nil)
	connectionBlockBody := connectionBlock.Body()

	connectionBlockBody.SetAttributeValue(defaults.Type, cty.StringVal(defaults.Ssh))
	connectionBlockBody.SetAttributeValue(defaults.User, cty.StringVal(terraformConfig.HarvesterConfig.HarvesterUser))

	hostExpression := defaults.Self + "." + defaults.PrivateIp
	host := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(hostExpression)},
	}

	connectionBlockBody.SetAttributeRaw(defaults.Host, host)

	keyPathExpression := defaults.File + `("` + terraformConfig.PrivateKeyPath + `")`
	keyPath := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(keyPathExpression)},
	}

	connectionBlockBody.SetAttributeRaw(defaults.PrivateKey, keyPath)
	connectionBlockBody.SetAttributeValue(defaults.Timeout, cty.StringVal(terraformConfig.HarvesterConfig.Timeout))
}
