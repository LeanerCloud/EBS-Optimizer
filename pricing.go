package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/pricing"
	"github.com/aws/aws-sdk-go-v2/service/pricing/types"
	"github.com/aws/aws-sdk-go/aws"
)

type Pricing struct {
	Product struct {
		ProductFamily string `json:"productFamily"`
		Attributes    struct {
			StorageMedia        string `json:"storageMedia"`
			MaxThroughputvolume string `json:"maxThroughputvolume"`
			VolumeType          string `json:"volumeType"`
			MaxIopsvolume       string `json:"maxIopsvolume"`
			Servicecode         string `json:"servicecode"`
			Usagetype           string `json:"usagetype"`
			LocationType        string `json:"locationType"`
			VolumeAPIName       string `json:"volumeApiName"`
			Location            string `json:"location"`
			Servicename         string `json:"servicename"`
			MaxVolumeSize       string `json:"maxVolumeSize"`
			Operation           string `json:"operation"`
			Group               string `json:"group"`
		} `json:"attributes"`
		Sku string `json:"sku"`
	} `json:"product"`
	ServiceCode string `json:"serviceCode"`
	Terms       struct {
		OnDemand struct {
			SKU struct {
				PriceDimensions struct {
					Dimension struct {
						Unit         string        `json:"unit"`
						EndRange     string        `json:"endRange"`
						Description  string        `json:"description"`
						AppliesTo    []interface{} `json:"appliesTo"`
						RateCode     string        `json:"rateCode"`
						BeginRange   string        `json:"beginRange"`
						PricePerUnit struct {
							USD string `json:"USD"`
						} `json:"pricePerUnit"`
					} `json:"Dimension"`
				} `json:"priceDimensions"`
				Sku            string    `json:"sku"`
				EffectiveDate  time.Time `json:"effectiveDate"`
				OfferTermCode  string    `json:"offerTermCode"`
				TermAttributes struct {
				} `json:"termAttributes"`
			} `json:"SKU"`
		} `json:"OnDemand"`
	} `json:"terms"`
	Version         string    `json:"version"`
	PublicationDate time.Time `json:"publicationDate"`
}

func getPricingData(filters []types.Filter) ([]Pricing, error) {

	var priceList []Pricing

	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-1"))
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}
	// Create a Pricing client with additional configuration
	svc := pricing.NewFromConfig(cfg)

	input := &pricing.GetProductsInput{
		ServiceCode: aws.String("AmazonEC2"),
		Filters:     filters,
	}

	paginator := pricing.NewGetProductsPaginator(svc, input)

	// Iterate through the Amazon S3 object pages.
	for paginator.HasMorePages() {
		// next page takes a context
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return nil, fmt.Errorf("failed to get a page, %w", err)
		}
		fmt.Println("page contents: ", page.PriceList)

		var p Pricing

		for _, item := range page.PriceList {

			dimension := regexp.MustCompile(`\"\w{16,}\.\w{10,}\.\w{10,}\"`)
			itemDimension := dimension.ReplaceAll([]byte(item), []byte("\"Dimension\""))

			sku := regexp.MustCompile(`\"\w{16,}\.\w{10,}\"`)
			itemSKU := sku.ReplaceAll(itemDimension, []byte("\"SKU\""))

			//	fmt.Printf("%v", string(itemSKU))

			err := json.Unmarshal(itemSKU, &p)
			if err != nil {
				fmt.Printf("error: ", err.Error())
			}
			priceList = append(priceList, p)
		}
	}
	return priceList, nil
}
