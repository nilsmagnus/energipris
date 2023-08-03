package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/guptarohit/asciigraph"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	prisDag := flag.String("prisdag", "idag", "Vise priser for i dag eller i morgen. Gyldige verdier: idag , imorgen")

	flag.Parse()

	if *prisDag != "idag" && *prisDag != "imorgen" {
		flag.Usage()
		log.Fatalf("Ugyldig prisdag: %s", *prisDag)
	}

	apiToken, hasToken := os.LookupEnv("TIBBER_API_TOKEN")

	if !hasToken {
		log.Fatal("TIBBER_API_TOKEN is not set.")
	}

	requestMap := map[string]string{"query": requestBody}
	requestJson, err := json.Marshal(requestMap)
	if err != nil {
		log.Fatalf("Klarte ikke lage json av requesten")
	}
	req, err := http.NewRequest("POST", "https://api.tibber.com/v1-beta/gql", bytes.NewBuffer(requestJson))
	if err != nil {
		log.Fatalf("Det er noe galt med requesten: %v", err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", apiToken))

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		log.Fatalf("Feil ved henting av data fra tibber: %v", err)
	}

	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatalf("Klarte ikke lese responsen fra tibber: %v", err)
	}

	tibberResponse := TibberResponse{}

	if err := json.Unmarshal(body, &tibberResponse); err != nil {
		log.Fatalf("Klarte ikke deserialisere responsen til tibber:%v \n %s", err, body)
	}

	if *prisDag == "idag" {
		data := mapTibberData(tibberResponse.Data.Viewer.Homes[0].CurrentSubscription.PriceInfo.Today)
		graph := asciigraph.Plot(data, asciigraph.Precision(3), asciigraph.Height(20), asciigraph.SeriesColors(asciigraph.Green))
		fmt.Println(graph)
	}
	if *prisDag == "imorgen" {
		data := mapTibberData(tibberResponse.Data.Viewer.Homes[0].CurrentSubscription.PriceInfo.Tomorrow)
		graph := asciigraph.Plot(data, asciigraph.Precision(3), asciigraph.Height(20), asciigraph.SeriesColors(asciigraph.Red))
		fmt.Println(graph)
	}

	fmt.Println("\t00:00\t\t\t06:00\t\t\t12:00\t\t\t18:00\t\t\t24:00")

}

func mapTibberData(days []Day) []float64 {
	result := make([]float64, len(days)*4)
	for i, f := range days {
		result[i*4] = f.Total
		result[i*4+1] = f.Total
		result[i*4+2] = f.Total
		result[i*4+3] = f.Total
	}
	return result
}

const requestBody = `
{
  viewer {
    homes {
     currentSubscription {
       priceInfo {
         current {
           total
           energy
           tax
           startsAt
         }
today {
            total
            energy
            tax
            startsAt
          }
          tomorrow {
            total
            energy
            tax
            startsAt
          }
       }

     }
   }
 }
}
`

type TibberResponse struct {
	Data struct {
		Viewer struct {
			Homes []struct {
				CurrentSubscription struct {
					PriceInfo struct {
						Current struct {
							Total    float64   `json:"total"`
							Energy   float64   `json:"energy"`
							Tax      float64   `json:"tax"`
							StartsAt time.Time `json:"startsAt"`
						} `json:"current"`
						Today    []Day `json:"today"`
						Tomorrow []Day `json:"tomorrow"`
					} `json:"priceInfo"`
				} `json:"currentSubscription"`
			} `json:"homes"`
		} `json:"viewer"`
	} `json:"data"`
}

type Day struct {
	Total    float64   `json:"total"`
	Energy   float64   `json:"energy"`
	Tax      float64   `json:"tax"`
	StartsAt time.Time `json:"startsAt"`
}
