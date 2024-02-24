import tensorflow as tf

def lambda_handler(event, context):
    # Load model
    model = tf.saved_model.load('simple_model')

    miles = event['miles']
    minutes = event['minutes']

    # Now return the prices
    prices = model(list(zip(miles, minutes)))[:, 0].numpy().tolist()
    
    print(prices)

    return {
        "statusCode": 200,
        "headers": {
            "Access-Control-Allow-Origin": "*"
        },
        "body": {
            "prices": prices
        }
    }