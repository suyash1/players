package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"models"
	"net/http"
	"sort"
	"strconv"
	"strings"
	_ "sync"
	"time"
	"utils"
)

type result struct {
	index  int
	status int
	res    http.Response
	err    error
}

func boundedParallelGet(urls []string, concurrencyLimit int) []result {

	// this buffered channel will block at the concurrency limit
	semaphoreChan := make(chan struct{}, concurrencyLimit)

	// this channel will not block and collect the http request results
	resultsChan := make(chan *result)

	// make sure we close these channels when we're done with them
	defer func() {
		close(semaphoreChan)
		close(resultsChan)
	}()

	// keep an index and loop through every url we will send a request to
	for i, url := range urls {

		// start a go routine with the index and url in a closure
		go func(i int, url string) {

			// this sends an empty struct into the semaphoreChan which
			// is basically saying add one to the limit, but when the
			// limit has been reached block until there is room
			semaphoreChan <- struct{}{}

			// send the request and put the response in a result struct
			// along with the index so we can sort them later along with
			// any error that might have occoured
			res := utils.GET(url)
			// if err != nil {
			// 	fmt.Println("error for url " + url)
			// 	fmt.Println(res.StatusCode)
			// }
			if res.StatusCode != 200 {
				fmt.Println(url, res.StatusCode)
			}
			result := &result{i, res.StatusCode, *res, nil}

			// now we can send the result struct through the resultsChan
			resultsChan <- result

			// once we're done it's we read from the semaphoreChan which
			// has the effect of removing one from the limit and allowing
			// another goroutine to start
			<-semaphoreChan

		}(i, url)
	}

	// make a slice to hold the results we're expecting
	var results []result

	// start listening for any results over the resultsChan
	// once we get a result append it to the result slice
	for {
		result := <-resultsChan
		results = append(results, *result)

		// if we've reached the expected amount of urls then stop
		if len(results) == len(urls) {
			break
		}
	}

	// let's sort these results real quick
	sort.Slice(results, func(i, j int) bool {
		return results[i].index < results[j].index
	})

	// now we're done we return the results
	return results
}

// we'll use the init function to set up the benchmark
// by making a slice of 100 URLs to send requets to
var urls []string
var endPoint = "https://vintagemonster.onefootball.com/api/teams/en/"

func init() {
	for i := 1; i <= 200; i++ {
		urls = append(urls, endPoint+strconv.Itoa(i)+".json")
	}
	//fmt.Println(urls)
}

var LookupTeams = map[string]bool{"Germany": false, "England": false, "France": false,
	"Spain": false, "Arsenal": false, "Chelsea": false, "Barcelona": false,
	}

func main() {
	benchmark := func(urls []string, concurrency int) string {
		startTime := time.Now()
		results := boundedParallelGet(urls, concurrency)
		teamPlayerMap := make(map[string][]models.Player)
		var playerData []models.Player
		for _, result := range results {
			data := make(map[string]interface{})
			if result.status == 200 {
				responseData, _ := ioutil.ReadAll(result.res.Body)
				if err := json.Unmarshal(responseData, &data); err != nil {
					fmt.Println(err)
				}
				innerData := data["data"].(map[string]interface{})
				teamData := innerData["team"].(map[string]interface{})
				teamName := teamData["name"].(string)
				if val, ok := LookupTeams[teamName]; ok {
					if false == val {
						teamPlayers := teamData["players"].([]interface{})
						// var players []models.Player
						for _, t := range teamPlayers {
							var teamPlayer map[string]interface{} = t.(map[string]interface{})
							tempPlayer := models.Player{}
							tempPlayer.FullName = teamPlayer["name"].(string)
							// tempPlayer.FirstName = teamPlayer["firstName"].(string)
							// tempPlayer.LastName = teamPlayer["lastName"].(string)
							age, _ := strconv.ParseInt(teamPlayer["age"].(string), 10, 64)
							tempPlayer.Age = age
							tempPlayer.TeamName = teamName
							playerData = append(playerData, tempPlayer)
						}
						teamPlayerMap[teamName] = playerData
						LookupTeams[teamName] = true
					}
				}

			}
		}
		// fmt.Println(playerData)
		playerTeamMap := make(map[string][]string)
		for _, data := range playerData {
			if val, ok := playerTeamMap[data.FullName]; ok {
				val = append(val, data.TeamName)
				playerTeamMap[data.FullName] = val
			} else {
				playerTeamMap[data.FullName + "; " + fmt.Sprint(data.Age) + "; "] = []string{data.TeamName}
			}
		}
		for k, v := range playerTeamMap {
			fmt.Println(k, strings.Join(v, ";"))
		}
		seconds := time.Since(startTime).Seconds()
		tmplate := "%d bounded parallel requests: %d/%d in %v"
		return fmt.Sprintf(tmplate, concurrency, len(results), len(urls), seconds)
	}
	fmt.Println(benchmark(urls, 100))
	// fmt.Println(benchmark(urls, 25))
	// fmt.Println(benchmark(urls, 50))
	// fmt.Println(benchmark(urls, 75))
	// fmt.Println(benchmark(urls, 100))
	//fmt.Println(benchmark(urls, 100))
}
