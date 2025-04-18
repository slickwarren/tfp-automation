package rke2k3s

import (
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	"github.com/zclconf/go-cty/cty"
)

// SetProxyConfig is a function that will set the proxy configurations in the main.tf file.
func SetProxyConfig(clusterBlockBody *hclwrite.Body, terraformConfig *config.TerraformConfig) error {
	agentVarsOneBlock := clusterBlockBody.AppendNewBlock(defaults.AgentEnvVars, nil)
	agentVarsOneBlockBody := agentVarsOneBlock.Body()

	agentVarsOneBlockBody.SetAttributeValue(defaults.ResourceName, cty.StringVal(httpProxy))
	agentVarsOneBlockBody.SetAttributeValue(defaults.Value, cty.StringVal("http://"+terraformConfig.Proxy.ProxyBastion+":3228"))

	agentVarsTwoBlock := clusterBlockBody.AppendNewBlock(defaults.AgentEnvVars, nil)
	agentVarsTwoBlockBody := agentVarsTwoBlock.Body()

	agentVarsTwoBlockBody.SetAttributeValue(defaults.ResourceName, cty.StringVal(httpsProxy))
	agentVarsTwoBlockBody.SetAttributeValue(defaults.Value, cty.StringVal("http://"+terraformConfig.Proxy.ProxyBastion+":3228"))

	agentVarsThreeBlock := clusterBlockBody.AppendNewBlock(defaults.AgentEnvVars, nil)
	agentVarsThreeBlockBody := agentVarsThreeBlock.Body()

	agentVarsThreeBlockBody.SetAttributeValue(defaults.ResourceName, cty.StringVal(noProxy))
	agentVarsThreeBlockBody.SetAttributeValue(defaults.Value, cty.StringVal(noProxyValue))

	return nil
}
