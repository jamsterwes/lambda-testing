package main

import (
	"context"
	"fmt"
	"math"

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
	model := tg.LoadModel("tf2_model", []string{"serve"}, nil)

	// Make tensor from input
	// TODO: dimension check
	input, _ := tf.NewTensor(event.Data)

	// Run model
	results := model.Exec([]tf.Output{
		model.Op("StatefulPartitionedCall", 0),
	}, map[tf.Output]*tf.Tensor{
		model.Op("serving_default_inputs", 0): input,
	})

	// Get prices
	modelResults := results[0]

	var modelResultsTensor [][]float32 = modelResults.Value().([][]float32)

	var prices []float32
	for _, result := range modelResultsTensor {
		// is this correct @nick?
		price := float64(result[0] + 5.75)
		price = math.Round(price+0.1) + 0.1 // round to nearest 0.9 (?)
		price = math.Max(8.87, price)

		prices = append(prices, float32(price))
	}

	return &PricesResponse{
		Prices: prices,
	}, nil
}

func main() {
	lambda.Start(HandleRequest)
}
