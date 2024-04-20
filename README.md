This program is used to auto transcode any uploads to HLS spec so that they can be streamed without issue
## ENV Variables Needed
- DB_ADDRESS - The http address of the MongoDB instance being used 
- MONGO_USERNAME - MongoDB username
- MONGO_PASS - MongoDB password
- DB_NAME - MongoDB database name
- BUCKET_NAME - Name of the S3 bucket being used
- AWS_ACCESS_KEY_ID - AWS Access Key that can reach the SQS Queue and S3 bucket 
- AWS_SECRET_ACCESS_KEY - Secret Key to match the S3 Access Key ID
- SQS_QUEUE_URL - URL of the SQS Queue being used
