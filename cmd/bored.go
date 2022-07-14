package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"

	"github.com/spf13/cobra"
)

//Constants
var base_url = "http://www.boredapi.com/api/activity/"
var type_values = []string{"education", "recreational", "social", "diy", "charity", "cooking", "relaxation", "music", "busywork"}

var accessMin_flag = "access-min"
var accessMax_flag = "access-max"
var type_flag = "type"
var participants_flag = "participants"
var priceMin_flag = "minprice"
var priceMax_flag = "maxprice"
var price_flag = "price"
var access_flag = "accessibility"

//struct for an idea
type Idea struct {
	Activity      string  `json:"activity"`
	Accessibility float64 `json:"accessibility"`
	Type          string  `json:"type"`
	Participants  int     `json:"participants"`
	Price         float64 `json:"price"`
	Link          string  `json:"link"`
	Key           string  `json:"key"`
}

//printing method
func (idea *Idea) print() {

	v := reflect.ValueOf(*idea)
	reflectType := v.Type()

	for i := 0; i < v.NumField(); i++ {
		if v.Field(i).Interface() != "" {
			fmt.Printf("%-13s = %v\n", reflectType.Field(i).Name, v.Field(i).Interface())
		}
	}

}

//when we send a bad get request we still
//obtain status 200 OK but with an error data
type BadResponse struct {
	Error string `json:"error"`
}

//here is some wrapper to elegantly handle response
type ResponseWrapper struct {
	Idea        *Idea        `json:""`
	BadResponse *BadResponse `json:""`
}

//method that helps unmarshall json to wrapper
func (wrapper *ResponseWrapper) UnmarshallJSON(dataBytes []byte) error {
	wrapperMap := make(map[string]interface{})
	var err error

	if err = json.Unmarshal(dataBytes, &wrapperMap); err != nil {
		return err
	}

	//if it contains bad response
	if _, ok := wrapperMap["error"]; ok {
		var badResponse BadResponse
		err = json.Unmarshal(dataBytes, &badResponse)

		if err != nil {
			return err
		}

		wrapper.BadResponse = &badResponse
	}

	//we can consider it as good response if it contains
	//one out of all idea's fields e.g. 'activity'
	if _, ok := wrapperMap["activity"]; ok {
		var idea Idea
		err = json.Unmarshal(dataBytes, &idea)

		if err != nil {
			return err
		}

		wrapper.Idea = &idea
	}

	return nil
}

// boredCmd represents the bored command
var boredCmd = &cobra.Command{
	Use:   "bored",
	Short: "Get a random idea",
	Long:  `This command fetches random idea what to do when you get bored from boredapi: https://www.boredapi.com/`,
	Run: func(cmd *cobra.Command, args []string) {
		handleFlags(cmd)
	},
}

func init() {
	rootCmd.AddCommand(boredCmd)

	boredCmd.PersistentFlags().Float32(access_flag, -1.0, "The exact factor describing how it is possible that an event has something to do with zero as most accessible, leave empty or -1.0 to get random, between: 0.0 - 1.0, can't be used with min and max argument")
	boredCmd.PersistentFlags().Float32(accessMin_flag, 0.0, "The minimal factor describing how it is possible that an event has something to do with zero as most accessible, default 0.0, between: 0.0 - 1.0")
	boredCmd.PersistentFlags().Float32(accessMax_flag, 1.0, "The maximal factor describing how it is possible that an event has something to do with zero as most accessible, default 1.0, between: 0.0 - 1.0")
	boredCmd.PersistentFlags().String(type_flag, "", "Type of the activity. Available types: education, recreational, social, diy, charity, cooking, relaxation, music, busywork")
	boredCmd.PersistentFlags().Int(participants_flag, -1, "The number of people that this activity could involve >= 0, leave empty or -1 for random")
	boredCmd.PersistentFlags().Float32(price_flag, -1.0, "Factor describing the exact cost of the event with zero being free, leave empty or -1.0 to get random, can't be used with min and max argument")
	boredCmd.PersistentFlags().Float32(priceMin_flag, 0.0, "Factor describing the minimum cost of the event with zero being free, default 0.0, between: 0.0 - 1.0")
	boredCmd.PersistentFlags().Float32(priceMax_flag, 1.0, "Factor describing the maximum cost of the event with zero being free, default 1.0, between: 0.0 - 1.0")
}

func handleFlags(cmd *cobra.Command) {
	access, _ := cmd.Flags().GetFloat32(access_flag)
	access_min, _ := cmd.Flags().GetFloat32(accessMin_flag)
	access_max, _ := cmd.Flags().GetFloat32(accessMax_flag)
	typeTerm, _ := cmd.Flags().GetString(type_flag)
	participants, _ := cmd.Flags().GetInt(participants_flag)
	price, _ := cmd.Flags().GetFloat32(price_flag)
	price_min, _ := cmd.Flags().GetFloat32(priceMin_flag)
	price_max, _ := cmd.Flags().GetFloat32(priceMax_flag)

	flagsMap := make(map[string]string)

	//ACCESS FLAG

	if (access < 0.0 || access > 1.0) && access != -1.0 {
		log.Printf("Invalid access argument, should be between 0.0 and 1.0 to get exact value or -1.0 / empty to get random, instead got: %g", access)
		return
	} else if access >= 0.0 && access <= 1.0 {
		flagsMap[access_flag] = fmt.Sprintf("%g", access)
	}

	//ACCESS_MIN FLAG

	if access_min > access_max {
		log.Printf("Invalid access_min and access_max arguments, access_min should be smaller than access_max, instead got: min: %g, max: %g", access_min, access_max)
		return
	}

	if access_min < 0.0 || access_min > 1.0 {
		log.Printf("Invalid access_min argument, should be between 0.0 and 1.0, instead got: %g", access_min)
		return
	}

	//we don't want to add an unnecessary argument, if it is equal to 0.0
	if access_min > 0.0 {
		flagsMap[accessMin_flag] = fmt.Sprintf("%g", access_min)
	}

	//ACCESS_MAX FLAG

	if access_max < 0.0 || access_max > 1.0 {
		log.Printf("Invalid access_max argument, should be between 0.0 and 1.0, instead got: %g", access_max)
		return
	}

	//we don't want to add an unnecessary argument, if it is equal to 1.0
	if access_max < 1.0 {
		flagsMap[accessMax_flag] = fmt.Sprintf("%g", access_max)
	}

	//TYPE FLAG

	if (!sliceContains(type_values, typeTerm)) && typeTerm != "" {
		log.Printf("Invalid type argument, should be one of available types: education, recreational, social, diy, charity, cooking, relaxation, music, busywork, instead got: %s", typeTerm)
		return
	} else if typeTerm != "" {
		flagsMap[type_flag] = typeTerm
	}

	if participants != -1 && participants < 0 {
		log.Printf("Invalid participants argument, should be >= 0 for exact value or left empty or -1 for random, instead got: %d", participants)
		return
	} else if participants > 0 {
		flagsMap[participants_flag] = fmt.Sprintf("%d", participants)
	}

	//PRICE FLAG

	if (price < 0.0 || price > 1.0) && price != -1.0 {
		log.Printf("Invalid price argument, should be between 0.0 and 1.0 to get exact value or -1.0 / empty to get random, instead got: %g", price)
		return
	} else if price >= 0.0 && price <= 1.0 {
		flagsMap[price_flag] = fmt.Sprintf("%g", price)
	}

	//PRICE_MIN FLAG

	if price_min > price_max {
		log.Printf("Invalid price_min and price_max arguments, price_min should be smaller than price_max, instead got: min: %g, max: %g", price_min, price_max)
		return
	}

	if price_min < 0.0 || price_min > 1.0 {
		log.Printf("Invalid price_min argument, should be between 0.0 and 1.0, instead got: %g", price_min)
		return
	}

	//we don't want to add an unnecessary argument, if it is equal to 0.0
	if price_min > 0.0 {
		flagsMap[priceMin_flag] = fmt.Sprintf("%g", price_min)
	}

	//PRICE_MAX FLAG

	if price_max < 0.0 || price_max > 1.0 {
		log.Printf("Invalid price_max argument, should be between 0.0 and 1.0, instead got: %g", price_max)
		return
	}

	//we don't want to add an unnecessary argument, if it is equal to 1.0
	if price_max < 1.0 {
		flagsMap[priceMax_flag] = fmt.Sprintf("%g", price_max)
	}

	getIdea(flagsMap)
}

func getIdea(endpoints map[string]string) {
	// wrapperMap := make(map[string]interface{})
	wrapper := ResponseWrapper{}

	responseBytes := getIdeaData(endpoints)

	if err := wrapper.UnmarshallJSON(responseBytes); err != nil {
		log.Printf("Could  not unmarshall response - %s", err.Error())
	}

	if wrapper.Idea != nil {
		fmt.Println()
		wrapper.Idea.print()
		fmt.Println()
	} else {
		fmt.Println(wrapper.BadResponse.Error)
	}
}

func getIdeaData(endpoints map[string]string) []byte {
	request, err := http.NewRequest(
		http.MethodGet,
		base_url,
		nil,
	)

	if err != nil {
		log.Printf("Could not request an idea - %s", err.Error())
	}

	query := request.URL.Query()

	//adding all additional arguments
	for key, value := range endpoints {
		query.Add(key, value)
	}
	request.URL.RawQuery = query.Encode()

	request.Header.Set("Content-Type", "application/json")

	response, err := http.DefaultClient.Do(request)

	if err != nil {
		log.Printf("Could not make a request - %s", err.Error())
	}

	responseBytes, err := ioutil.ReadAll(response.Body)

	if err != nil {
		log.Printf("Could not read a request body - %s", err.Error())
	}

	return responseBytes
}

func sliceContains(slice []string, value string) bool {
	for _, word := range slice {
		if word == value {
			return true
		}
	}
	return false
}
