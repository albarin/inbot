package main

import (
	"context"
	"fmt"
	"os"

	"github.com/albarin/indexa/pkg/indexa"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

// Response is of type APIGatewayProxyResponse since we're leveraging the
// AWS Lambda Proxy Request functionality (default behavior)
//
// https://serverless.com/framework/docs/providers/aws/events/apigateway/#lambda-proxy-integration
type Response events.APIGatewayProxyResponse

const indexaURL = "https://api.indexacapital.com"

func getIndexaPerformace() (*indexa.Performance, error) {
	c := indexa.NewIndexaClient(indexaURL, os.Getenv("TOKEN"))

	me, err := c.Me()
	if err != nil {
		return nil, fmt.Errorf("error: %", err)
	}

	p, err := c.Performance(me.Accounts[0].AccountNumber)
	if err != nil {
		return nil, fmt.Errorf("error: %", err)
	}

	return p, nil
}

// Handler is our lambda handler invoked by the `lambda.Start` function call
func Handler(ctx context.Context) (Response, error) {
	p, err := getIndexaPerformace()
	if err != nil {
		return Response{StatusCode: 500}, nil
	}

	color := "#47bc2d"
	if p.Return.TimeReturn < 0 {
		color = "#bc2d2d"
	}

	message := fmt.Sprintf(`{
		"response_type": "ephemeral",
		"attachments": [
			{
				"color": "%s",
				"fields": [
					{
						"title": "Rentabilidad por dinero",
						"value": "%.1f%% acumulada (%.1f%% TAE)",
						"short": false
					},
					{
						"title": "Rentabilidad por tiempo",
						"value": "%.1f%% acumulada (%.1f%% TAE)",
						"short": false
					}
				]
			},
			{
				"color": "#804de1",
				"fields": [
					{
						"title": "Rentabilidad ayer",
						"value": "%.1f%%",
						"short": false
					}
				]
			},
			{
				"color": "#96beff",
				"fields": [
					{
						"title": "Aportaciones",
						"value": "%.2f€",
						"short": false
					},
					{
						"title": "Rentabilidad",
						"value": "%.2f€",
						"short": false
					},
					{
						"title": "Valor total",
						"value": "%.2f€",
						"short": false
					},
					{
						"title": "Volatilidad",
						"value": "%.2f%%",
						"short": false
					}
				]
			}
		]
	}`, color,
		p.Return.TimeReturn*100,
		p.Return.TimeReturnAnnual*100,
		p.Return.MoneyReturn*100,
		p.Return.MoneyReturnAnnual*100,
		p.Return.MoneyReturnAnnual*100,
		float64(p.Return.Investment),
		p.Return.Pl,
		p.Return.TotalAmount,
		p.Volatility*100,
	)

	resp := Response{
		StatusCode:      200,
		IsBase64Encoded: false,
		Body:            message,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}

	return resp, nil
}

func main() {
	lambda.Start(Handler)
}
