package harvester

import (
	"fmt"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	"github.com/zclconf/go-cty/cty"
)

const (
	eachValue       = "each.value"
	forEach         = "for_each"
	rke2InstanceIDs = "rke2_instance_ids"
)

// CreateTargetGroupAttachments is a function that will set the target group attachments configurations in the main.tf file.
func CreateTargetGroupAttachments(rootBody *hclwrite.Body, lbTargetGroupAttachment, targetGroupAttachmentServer string, port int64) {
	targetGroupBlock := rootBody.AppendNewBlock(defaults.Resource, []string{lbTargetGroupAttachment, targetGroupAttachmentServer})
	targetGroupBlockBody := targetGroupBlock.Body()

	instanceValueExpression := defaults.Local + "." + rke2InstanceIDs
	values := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(instanceValueExpression)},
	}

	targetGroupBlockBody.SetAttributeRaw(forEach, values)

	targetGroupExpression := defaults.LoadBalancerTargetGroup + "." + defaults.TargetGroupPrefix + fmt.Sprint(port) + ".arn"
	values = hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(targetGroupExpression)},
	}

	targetGroupBlockBody.SetAttributeRaw(defaults.TargetGroupARN, values)

	targetIDExpression := eachValue
	values = hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(targetIDExpression)},
	}

	targetGroupBlockBody.SetAttributeRaw(defaults.TargetID, values)
	targetGroupBlockBody.SetAttributeValue(defaults.Port, cty.NumberIntVal(port))
}

// CreateInternalTargetGroupAttachments is a function that will set the internal target group attachments configurations in the main.tf file.
func CreateInternalTargetGroupAttachments(rootBody *hclwrite.Body, lbTargetGroupAttachment, targetGroupAttachmentServer string, port int64) {
	targetGroupBlock := rootBody.AppendNewBlock(defaults.Resource, []string{lbTargetGroupAttachment, targetGroupAttachmentServer})
	targetGroupBlockBody := targetGroupBlock.Body()

	instanceValueExpression := defaults.Local + "." + rke2InstanceIDs
	values := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(instanceValueExpression)},
	}

	targetGroupBlockBody.SetAttributeRaw(forEach, values)

	targetGroupExpression := defaults.LoadBalancerTargetGroup + "." + defaults.TargetGroupInternalPrefix + fmt.Sprint(port) + ".arn"
	values = hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(targetGroupExpression)},
	}

	targetGroupBlockBody.SetAttributeRaw(defaults.TargetGroupARN, values)

	targetIDExpression := eachValue
	values = hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(targetIDExpression)},
	}

	targetGroupBlockBody.SetAttributeRaw(defaults.TargetID, values)
	targetGroupBlockBody.SetAttributeValue(defaults.Port, cty.NumberIntVal(port))
}
