package harvester

import (
	"fmt"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/format"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	"github.com/zclconf/go-cty/cty"
)

// CreateWindowsHarvesterInstances is a function that will set the Windows Harvester instances configurations in the main.tf file.
func CreateWindowsHarvesterInstances(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, terratestConfig *config.TerratestConfig,
	hostnamePrefix string) {
	configBlock := rootBody.AppendNewBlock(defaults.Resource, []string{defaults.AwsInstance, hostnamePrefix + "-windows"})
	configBlockBody := configBlock.Body()

	configBlockBody.SetAttributeValue(defaults.Count, cty.NumberIntVal(terratestConfig.WindowsNodeCount))

	configBlockBody.SetAttributeValue(defaults.Ami, cty.StringVal(terraformConfig.HarvesterConfig.WindowsAMI))
	configBlockBody.SetAttributeValue(defaults.InstanceType, cty.StringVal(terraformConfig.HarvesterConfig.WindowsInstanceType))
	configBlockBody.SetAttributeValue(defaults.SubnetId, cty.StringVal(terraformConfig.HarvesterConfig.HarvesterSubnetID))

	securityGroups := format.ListOfStrings(terraformConfig.HarvesterConfig.HarvesterSecurityGroups)
	configBlockBody.SetAttributeRaw(defaults.VpcSecurityGroupIds, securityGroups)
	configBlockBody.SetAttributeValue(defaults.KeyName, cty.StringVal(terraformConfig.HarvesterConfig.WindowsKeyName))

	configBlockBody.AppendNewline()

	rootBlockDevice := configBlockBody.AppendNewBlock(defaults.RootBlockDevice, nil)
	rootBlockDeviceBody := rootBlockDevice.Body()

	rootBlockDeviceBody.SetAttributeValue(defaults.VolumeSize, cty.NumberIntVal(terraformConfig.HarvesterConfig.HarvesterRootSize))

	configBlockBody.AppendNewline()

	tagsBlock := configBlockBody.AppendNewBlock(defaults.Tags+" =", nil)
	tagsBlockBody := tagsBlock.Body()

	expression := fmt.Sprintf(`"%s-windows-${`+defaults.Count+`.`+defaults.Index+`}"`, terraformConfig.ResourcePrefix)
	tags := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(expression)},
	}

	tagsBlockBody.SetAttributeRaw(defaults.Name, tags)

	configBlockBody.AppendNewline()

	connectionBlock := configBlockBody.AppendNewBlock(defaults.Connection, nil)
	connectionBlockBody := connectionBlock.Body()

	connectionBlockBody.SetAttributeValue(defaults.Type, cty.StringVal(defaults.Ssh))
	connectionBlockBody.SetAttributeValue(defaults.User, cty.StringVal(terraformConfig.HarvesterConfig.WindowsHarvesterUser))

	hostExpression := defaults.Self + "." + defaults.PublicIp
	host := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(hostExpression)},
	}

	connectionBlockBody.SetAttributeRaw(defaults.Host, host)
	connectionBlockBody.SetAttributeValue(defaults.TargetPlatform, cty.StringVal(defaults.Windows))

	keyPathExpression := defaults.File + `("` + terraformConfig.WindowsPrivateKeyPath + `")`
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

// CreateAirgappedWindowsHarvesterInstances is a function that will set the Windows Harvester instances configurations in the main.tf file.
func CreateAirgappedWindowsHarvesterInstances(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, hostnamePrefix string) {
	configBlock := rootBody.AppendNewBlock(defaults.Resource, []string{defaults.AwsInstance, hostnamePrefix})
	configBlockBody := configBlock.Body()

	configBlockBody.SetAttributeValue(defaults.AssociatePublicIPAddress, cty.BoolVal(false))
	configBlockBody.SetAttributeValue(defaults.Ami, cty.StringVal(terraformConfig.HarvesterConfig.WindowsAMI))
	configBlockBody.SetAttributeValue(defaults.InstanceType, cty.StringVal(terraformConfig.HarvesterConfig.WindowsInstanceType))
	configBlockBody.SetAttributeValue(defaults.SubnetId, cty.StringVal(terraformConfig.HarvesterConfig.HarvesterSubnetID))

	securityGroups := format.ListOfStrings(terraformConfig.HarvesterConfig.HarvesterSecurityGroups)
	configBlockBody.SetAttributeRaw(defaults.VpcSecurityGroupIds, securityGroups)
	configBlockBody.SetAttributeValue(defaults.KeyName, cty.StringVal(terraformConfig.HarvesterConfig.WindowsKeyName))

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
	connectionBlockBody.SetAttributeValue(defaults.User, cty.StringVal(terraformConfig.HarvesterConfig.WindowsHarvesterUser))

	hostExpression := defaults.Self + "." + defaults.PrivateIp
	host := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(hostExpression)},
	}

	connectionBlockBody.SetAttributeRaw(defaults.Host, host)

	keyPathExpression := defaults.File + `("` + terraformConfig.WindowsPrivateKeyPath + `")`
	keyPath := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(keyPathExpression)},
	}

	connectionBlockBody.SetAttributeRaw(defaults.PrivateKey, keyPath)
	connectionBlockBody.SetAttributeValue(defaults.Timeout, cty.StringVal(terraformConfig.HarvesterConfig.Timeout))
}
