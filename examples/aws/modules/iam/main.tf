# Copyright 2022 Z5Labs and Contributors
# 
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
# 
#     http://www.apache.org/licenses/LICENSE-2.0
# 
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

terraform {
  required_providers {
    aws = {
      source = "hashicorp/aws"
      version = ">= 4.34.0"
    }
  }
}

resource "aws_iam_role" "evrys_execution_role" {
  name               = "evrys-execution-role"
  assume_role_policy = data.aws_iam_policy_document.fargate_assume_role_policy.json
  managed_policy_arns = [
    aws_iam_policy.dynamodb_policy.arn,
    aws_iam_policy.sns_policy.arn
  ]
}

resource "aws_iam_policy" "dynamodb_policy" {
  name = "dynamodb-policy"
  policy = data.aws_iam_policy_document.dynamodb_policy.json
}

resource "aws_iam_policy" "sns_policy" {
  name = "sns-policy"
  policy = data.aws_iam_policy_document.sns_policy.json
}

data "aws_iam_policy_document" "fargate_assume_role_policy" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["ecs-tasks.amazonaws.com"]
    }
  }
}

data "aws_iam_policy_document" "dynamodb_policy" {
  statement {
    actions = [
      "dynamodb:PutItem",
      "dynamodb:Get*",
      "dynamodb:Query",
      "dynamodb:Scan",
    ]

    resources = [
      var.dynamodb_table_arn
    ]
  }
}

data "aws_iam_policy_document" "sns_policy" {
  statement {
    actions = [
      "sns:Publish"
    ]

    resources = [
      var.sns_topic_arn
    ]
  }
}