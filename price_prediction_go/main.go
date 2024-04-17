package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-lambda-go/lambda"

	tf "github.com/galeone/tensorflow/tensorflow/go"
	tg "github.com/galeone/tfgo"
)

type PricesRequest struct {
	Data [][]float32 `json:"data"`
}

type PricesResponse struct {
	Prices []float32 `json:"prices"`
}

func HandleRequest(ctx context.Context, event *PricesRequest) (*PricesResponse, error) {
	if event == nil {
		return nil, fmt.Errorf("received nil event")
	}

	// Load model
	model := tg.LoadModel("tf_model", []string{"serve"}, nil)

	// Make tensor from input
	// TODO: dimension check
	input, _ := tf.NewTensor(event.Data)

	// Run model
	results := model.Exec([]tf.Output{
		model.Op("StatefulPartitionedCall", 0),
	}, map[tf.Output]*tf.Tensor{
		model.Op("serving_default_normalization_4_input", 0): input,
	})

	// Get prices
	modelResults := results[0]

	var modelResultsTensor [][]float32 = modelResults.Value().([][]float32)

	var prices []float32
	for _, result := range modelResultsTensor {
		prices = append(prices, result...)
	}

	return &PricesResponse{
		Prices: prices,
	}, nil
}

func main() {
	lambda.Start(HandleRequest)
}
