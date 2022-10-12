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

resource "aws_sns_topic" "topic" {
  name = var.topic_name
}

resource "aws_sqs_queue" "topic_queue" {
  name = var.topic_name
}

resource "aws_sns_topic_subscription" "topic_to_sqs" {
  topic_arn = aws_sns_topic.topic.arn
  protocol  = "sqs"
  endpoint  = aws_sqs_queue.topic_queue.arn
}