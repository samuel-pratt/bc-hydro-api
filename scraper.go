package main

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type Response struct {
	Regions   []Region  `json:"regions"`
	ScrapedAt time.Time `json:"scrapedAt"`
}

type Region struct {
	Name        string   `json:"name"`
	OutageCount int      `json:"outagesCount"`
	Outages     []Outage `json:"outages"`
}

type Outage struct {
	Municipality      string `json:"municipality"`
	Time              string `json:"time"`
	Status            string `json:"status"`
	Area              string `json:"area"`
	CustomersAffected int    `json:"customersAffected"`
	Cause             string `json:"cause"`
	LastUpdated       string `json:"lastUpdated"`
	Map               string `json:"mapLink"`
}

func ScrapeOutages() Response {
	bcHydroLink := "https://www.bchydro.com/power-outages/app/outage-list.html"

	response, err := http.Get(bcHydroLink)
	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()

	document, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		log.Fatal("Error loading HTTP response body. ", err)
	}

	var regions []Region

	document.Find(".outage-list-details").Each(func(index int, outageInfo *goquery.Selection) {
		region := Region{
			Name:        "",
			OutageCount: 0,
			Outages:     nil,
		}

		outageInfo.Find(".col-1").Each(func(indextr int, regionInfo *goquery.Selection) {
			regionName := regionInfo.Find("b").First()
			region.Name = strings.TrimSpace(regionName.Text())
		})

		outageInfo.Find(".municipality-list").Each(func(indextr int, table *goquery.Selection) {
			table.Find("tbody").Each(func(indextr int, tbody *goquery.Selection) {
				var outages []Outage

				tbody.Find("tr").Each(func(indextr int, tr *goquery.Selection) {
					var outage Outage

					municipality := tr.Find(".municip").First()
					outage.Municipality = municipality.Text()

					offSince := tr.Find(".off-since").First()
					offSinceJsString := strings.TrimSpace(offSince.Text())
					offSinceJsString = strings.Replace(offSinceJsString, "document.write(format_date(new Date('", "", 1)
					offSinceJsString = strings.Replace(offSinceJsString, "')));", "", 1)
					outage.Time = offSinceJsString

					status := tr.Find(".status").First()
					statusString := strings.TrimSpace(status.Text())
					statusString = strings.Replace(statusString, "\t", "", -1)
					statusString = strings.Replace(statusString, "\n", "", -1)
					statusString = strings.Replace(statusString, "document.write(format_date(new Date('", "", 1)
					statusString = strings.Replace(statusString, "')));", "", 1)
					outage.Status = statusString

					area := tr.Find(".area").First()
					outage.Area = strings.Split(area.Text(), "\n")[0]

					mapLink := area.Find("a").First()
					outage.Map = "https://www.bchydro.com/power-outages/app/" + mapLink.AttrOr("href", "")

					customersAffected := tr.Find(".cust-aff").First()
					if strings.Contains(customersAffected.Text(), "< ") {
						outage.CustomersAffected, err = strconv.Atoi(strings.Split(customersAffected.Text(), "< ")[1])
					} else {
						outage.CustomersAffected, err = strconv.Atoi(customersAffected.Text())
					}

					cause := tr.Find(".cause").First()
					outage.Cause = cause.Text()

					lastUpdated := tr.Find(".last-updated").First()
					lastUpdatedJsString := strings.TrimSpace(lastUpdated.Text())
					lastUpdatedJsString = strings.Replace(lastUpdatedJsString, "document.write(format_date(new Date('", "", 1)
					lastUpdatedJsString = strings.Replace(lastUpdatedJsString, "')));", "", 1)
					outage.LastUpdated = lastUpdatedJsString

					outages = append(outages, outage)
				})

				region.Outages = outages
			})
		})

		region.OutageCount = len(region.Outages)

		regions = append(regions, region)
	})

	// Add timestamp to data
	currentTime := time.Now()

	// Add schedule and timestamp to response object
	scraperResponse := Response{
		Regions:   regions,
		ScrapedAt: currentTime,
	}

	return scraperResponse
}
