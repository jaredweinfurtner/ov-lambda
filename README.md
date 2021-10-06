# OV-Lambda

An AWS Lambda function that checks my nearest movie theater in Germany for which movie is currently being presented in
OV (Original Version) and notifying me by email when it changes.

More importantly, it showcases how you can integrate other AWS services (DynamoDB, Simple Email Service, etc) with
Lambda expressions in golang.

## Usage

I assume you already have an AWS account, so we'll just go straight into it:

1. First you'll need to create an [AWS DynamoDB](https://aws.amazon.com/dynamodb/) table with the following structure

   | Attribute | Type | 
   |---|---|
   | id (partition key) | String |
   | url | String |
   | movie | String |

   and insert the following row as an example:

   | Attribute | Value | 
   |---|---|
   | id | Esslingen |
   | url | https://esslingen.traumpalast.de/index.php/PID/138/R/70.html |
   | movie | \<empty\> |

2. Next, you'll need to register for [AWS Simple Email Service (SES)](https://aws.amazon.com/ses/) with your custom
   domain to send notification emails
3. Finally, create your [AWS Lambda](https://aws.amazon.com/lambda/) function. AWS Lambda allows a zipped golang binary
   to be uploaded manually. To generate this, simply call `make all` and it will generate `ov-lambda.zip` to upload.

   You will need to set the AWS Lambda environment variables:

   | Variable | Description |
       |---|---|
   | DB_TABLE_NAME | your dynamodb table name |
   | THEATER | The theater table key in your dynamodb |
   | EMAIL_CHARSET | The SES email character set (UTF-8) |
   | FROM_EMAIL | The SES email from address |
   | TO_EMAIL | The SES email to address (recipient of notification)|

   It is recommended that you use an AWS [EventBridge](https://aws.amazon.com/eventbridge/) trigger for your lambda so
   that it can poll the URL every hour (or whatever cron express you want)

## License

OV-Lambda is open-sourced under the Unlicense. See the [LICENSE](./LICENSE) file for details.
