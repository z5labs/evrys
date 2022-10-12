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
      version = "4.34.0"
    }
  }
}

provider "aws" {
  shared_config_files      = [var.config_file]
  shared_credentials_files = [var.credentials_file]
  profile                  = var.profile
}

module "evrys_dynamodb" {
  source = "./modules/dynamo"
  providers = {
    aws = aws
  }

  # variables
  table_name = "evrys-events"
}

module "evrys_sns2sqs" {
  source = "./modules/sns2sqs"
  providers = {
    aws = aws
  }

  # variables
  topic_name = "evrys-notifications"
}

module "evrys_iam" {
  source = "./modules/iam"
  providers = {
    aws = aws
  }

  # variables
  dynamodb_table_arn = module.evrys_dynamodb.table_arn
  sns_topic_arn = module.evrys_sns2sqs.topic_arn
}
