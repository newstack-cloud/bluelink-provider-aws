# Complete IAM Managed Policy

This example creates a comprehensive IAM managed policy with multiple statements and all optional fields.

```javascript
{
  "policyName": "EC2FullAccessPolicy",
  "policyDocument": {
    "Version": "2012-10-17",
    "Statement": [
      {
        "Effect": "Allow",
        "Action": [
          "ec2:DescribeInstances",
          "ec2:DescribeSecurityGroups",
          "ec2:DescribeVpcs",
          "ec2:DescribeSubnets",
          "ec2:DescribeRouteTables",
          "ec2:DescribeInternetGateways",
          "ec2:DescribeNatGateways",
          "ec2:DescribeVpcEndpoints",
          "ec2:DescribeVpcPeeringConnections",
          "ec2:DescribeTransitGateways",
          "ec2:DescribeTransitGatewayVpcAttachments",
          "ec2:DescribeTransitGatewayRouteTables",
          "ec2:DescribeTransitGatewayMulticastDomains",
          "ec2:DescribeTransitGatewayPeeringAttachments",
          "ec2:DescribeTransitGatewayConnects",
          "ec2:DescribeTransitGatewayConnectPeers",
          "ec2:DescribeTransitGatewayPrefixListReferences",
          "ec2:DescribeTransitGatewayRouteTableAnnouncements",
          "ec2:DescribeTransitGatewayAttachments",
          "ec2:DescribeTransitGatewayRouteTables",
          "ec2:DescribeTransitGatewayMulticastDomains",
          "ec2:DescribeTransitGatewayPeeringAttachments",
          "ec2:DescribeTransitGatewayConnects",
          "ec2:DescribeTransitGatewayConnectPeers",
          "ec2:DescribeTransitGatewayPrefixListReferences",
          "ec2:DescribeTransitGatewayRouteTableAnnouncements"
        ],
        "Resource": "*"
      },
      {
        "Effect": "Allow",
        "Action": [
          "ec2:RunInstances",
          "ec2:StartInstances",
          "ec2:StopInstances",
          "ec2:TerminateInstances",
          "ec2:RebootInstances",
          "ec2:CreateSecurityGroup",
          "ec2:CreateTags",
          "ec2:DeleteTags",
          "ec2:ModifyInstanceAttribute",
          "ec2:ModifySecurityGroupRules",
          "ec2:AuthorizeSecurityGroupIngress",
          "ec2:RevokeSecurityGroupIngress",
          "ec2:AuthorizeSecurityGroupEgress",
          "ec2:RevokeSecurityGroupEgress"
        ],
        "Resource": "*",
        "Condition": {
          "StringEquals": {
            "aws:RequestTag/Environment": "Production"
          }
        }
      }
    ]
  },
  "description": "Policy that provides full access to EC2 resources with conditions",
  "path": "/managed-policies/",
  "tags": [
    {
      "key": "Environment",
      "value": "Production"
    },
    {
      "key": "Project",
      "value": "EC2Management"
    },
    {
      "key": "Owner",
      "value": "DevOps Team"
    },
    {
      "key": "CostCenter",
      "value": "12345"
    },
    {
      "key": "Compliance",
      "value": "SOX"
    }
  ]
}
``` 