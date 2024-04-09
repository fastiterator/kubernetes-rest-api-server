# IAM for this application
The IAM user I used to put this together definitely has more
privileges than absolutely needed, though many unexpected policies
were required to make this setup work.  Policies attached to this user
are as follows
## AWS Managed
AmazonEC2FullAccess
AmazonEKS_CNI_Policy
AmazonEKSClusterPolicy
AmazonEKSFargatePodExecutionRolePolicy
AmazonEKSLocalOutpostClusterPolicy
AmazonEKSServicePolicy
AmazonEKSVPCResourceController
AmazonEKSWorkerNodePolicy
AWSCloudFormationFullAccess
AWSFaultInjectionSimulatorEKSAccess
CloudWatchFullAccess
## Self-Managed

### account-IAM-user-access-keys
```
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "CreateOwnAccessKeys",
            "Effect": "Allow",
            "Action": [
                "iam:GetUser",
                "iam:CreateAccessKey",
                "iam:ListAccessKeys"
            ],
            "Resource": "arn:aws:iam::*:user/${aws:username}"
        },
        {
            "Sid": "ManageOwnAccessKeys",
            "Effect": "Allow",
            "Action": [
                "iam:CreateAccessKey",
                "iam:DeleteAccessKey",
                "iam:GetAccessKeyLastUsed",
                "iam:GetUser",
                "iam:ListAccessKeys",
                "iam:UpdateAccessKey",
                "iam:ListServiceSpecificCredentials"
            ],
            "Resource": "arn:aws:iam::*:user/${aws:username}"
        },
        {
            "Sid": "ListUsers",
            "Effect": "Allow",
            "Action": "iam:ListUsers",
            "Resource": "*"
        }
    ]
}
```

### account-IAM-user-password-change
```
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "ViewAccountPasswordRequirements",
            "Effect": "Allow",
            "Action": "iam:GetAccountPasswordPolicy",
            "Resource": "*"
        },
        {
            "Sid": "ChangeOwnPassword",
            "Effect": "Allow",
            "Action": [
                "iam:GetLoginProfile",
                "iam:GetUser",
                "iam:ChangePassword"
            ],
            "Resource": "arn:aws:iam::*:user/${aws:username}"
        },
        {
            "Sid": "SigningCertsRO",
            "Effect": "Allow",
            "Action": [
                "iam:ListSigningCertificates"
            ],
            "Resource": "arn:aws:iam::*:user/${aws:username}"
        }
    ]
}
```

### ec2-describe-all
```
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": "ec2:Describe*",
            "Resource": "*"
        }
    ]
}
```

### eks-admin-all
```
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "VisualEditor0",
            "Effect": "Allow",
            "Action": "eks:*",
            "Resource": "*"
        }
    ]
}
```
