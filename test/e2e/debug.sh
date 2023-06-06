#!/bin/bash

set -x

echo "Hostname: "
hostname
echo "Hostname IP: "
hostname -I

echo "Checking aws-cli version..."
aws --version

echo "Checking aws sts-get-caller-identity"
aws sts get-caller-identity

echo "Checking k8s svc"
kubectl get svc --v=9

echo "Checking EKS API server availability"
aws eks describe-cluster --name e2eci-${CI_PIPELINE_ID}-${CI_PROJECT_ID}-eks-e2e --region us-east-1 --query cluster.resourcesVpcConfig

echo "Debug AWS CNI"
/bin/bash /opt/cni/bin/aws-cni-support.sh

echo "Debug cluster security groups"
aws eks describe-cluster --name e2eci-${CI_PIPELINE_ID}-${CI_PROJECT_ID}-eks-e2e --query cluster.resourcesVpcConfig.securityGroupIds

echo "Get VPC ID"
vpc_id=$(aws eks describe-cluster --name e2eci-${CI_PIPELINE_ID}-${CI_PROJECT_ID}-eks-e2e --query cluster.resourcesVpcConfig.vpcId --output text)

echo "Debug cluster security groups",
aws ec2 describe-security-groups --filters \"Name=vpc-id,Values=${vpc_id}\" --query "SecurityGroups[*].GroupId"
