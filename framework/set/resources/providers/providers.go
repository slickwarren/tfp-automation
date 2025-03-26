package providers

import (
	"fmt"
	"os"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	defaultproviders "github.com/rancher/tfp-automation/defaults/providers"
	"github.com/rancher/tfp-automation/framework/set/resources/providers/aws"
	"github.com/rancher/tfp-automation/framework/set/resources/providers/harvester"
)

type ProviderResourceFunc func(file *os.File, newFile *hclwrite.File, tfBlockBody, rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig,
	terratestConfig *config.TerratestConfig, instances []string) (*os.File, error)

type ProviderResources struct {
	CreateAirgap    ProviderResourceFunc
	CreateNonAirgap ProviderResourceFunc
}

func TunnelToProvider(provider string) ProviderResources {
	switch provider {
	case defaultproviders.Harvester:
		return ProviderResources{
			CreateAirgap:    harvester.CreateAirgappedHarvesterResources,
			CreateNonAirgap: harvester.CreateHarvesterResources,
		}

	case defaultproviders.AWS:
		return ProviderResources{
			CreateAirgap:    aws.CreateAirgappedAWSResources,
			CreateNonAirgap: aws.CreateAWSResources,
		}
	// case defaultproviders.Linode:
	// 	return ProviderResources{
	// 		CreateAirgap:    linode.CreateAirgappedAWSResources,
	// 		CreateNonAirgap: linode.CreateAWSResources,
	// 	}
	default:
		panic(fmt.Sprintf("Provider %v not found", provider))
	}

}
