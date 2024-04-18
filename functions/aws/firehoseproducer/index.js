const AWSXRay = require('aws-xray-sdk')
const AWS = AWSXRay.captureAWS(require('aws-sdk'))

const firehose = new AWS.Firehose({region: "eu-central-1"});
const firehoseName = process.env.firehose_name;

exports.handler = async (event, context, callback) => {

    return new Promise(function (resolve, reject) {

        const params = {
            DeliveryStreamName: firehoseName,
            Record: {
                Data: JSON.stringify(JSON.parse(event["body"]))
            }
        };

        firehose.putRecord(params, function (err, data) {
            if (err) console.log(err, err.stack); // an error occurred
            else console.log('Firehose Successful', data);           //         successful response
        });
    });
};