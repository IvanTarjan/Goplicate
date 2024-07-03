package goplicate

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)
// While this generic struct can be used by itself, it is recommended to make a package with a concrete implementation.
//
// The replicateConsumer struct is a Generic consumer that receives makes requests to the replicate api using the Input and Output types provided.
//
// The version attribute is optional, if it is not needed leave it as an empty string.
type ReplicateConsumer[Input, Output any] struct {
	ApiKey              string
	CreatePredictionUrl string
	Version             string
	BasePredictionRequest   *predictionRequest[Input]
	CheckStatusFinished func(output *Output) bool
	ExtractData         func(output *Output) (string, error)
}

// While this generic struct can be used by itself, it is recommended to make a package with a concrete implementation.
//
// The replicateConsumer struct is a Generic consumer that receives makes requests to the replicate api using the Input and Output types provided.
//
// The version attribute is optional, if it is not needed leave it as an empty string.
//
//The extractData func is needed to extract the string from the Generic Output that the request returns. 
//As is noticed that a lot of models returned the same output except for the input field, I added a default GetOutput func.
//The same applies to the CheckStatusFinished func
func NewReplicateConsumer[Input, Output any](apiKey, createPredictionUrl, version string, baseInput Input, extractData func(output *Output) (string, error), checkStatusFinished func(output *Output) bool) *ReplicateConsumer[Input, Output] {
	return &ReplicateConsumer[Input, Output]{
		ApiKey:              apiKey,
		CreatePredictionUrl: createPredictionUrl,
		Version:             version,
		BasePredictionRequest: &predictionRequest[Input]{
			Version: version,
			Stream:  false,
			Input:   baseInput,
		},
		ExtractData:         extractData,
		CheckStatusFinished: checkStatusFinished,
	}
}

// Create prediction receives a function that allows you to customize only the desired fields of the input without changing the rest of the prediction request and input.
//
// The function returns a prediction response that contains the urls needed to get the output of the prediction.
func (r *ReplicateConsumer[Input, Output]) CreatePrediction(customizeInput func(input Input) Input) (*predictionResponse[Input], error) {
	customizedPredictionRequest := r.BasePredictionRequest
	customizedPredictionRequest.Input = customizeInput(r.BasePredictionRequest.Input)
	client := &http.Client{}
	requestBodyJson, err := json.Marshal(customizedPredictionRequest)
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequest("POST", r.CreatePredictionUrl, bytes.NewBuffer(requestBodyJson))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Authorization", "Bearer "+r.ApiKey)
	request.Header.Set("Content-Type", "application/json")
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	prediction := &predictionResponse[Input]{}
	err = json.NewDecoder(response.Body).Decode(prediction)
	if err != nil {
		return nil, err
	}
	return prediction, nil
}

//GetOutput receives the url of the output and channels to send the output and errors
//
//The function is meant to be run as a goroutine to be as similar as possible to the replicateStreamConsumer. 
func (r *ReplicateConsumer[Input, Output]) GetOutput(outputUrl string, outputChan chan string, errorChan chan error) {
	client := http.Client{}
	request, err := http.NewRequest("GET", outputUrl, nil)
	if err != nil {
		fmt.Println("error creating request")
		errorChan <- err
		return
	}
	defer func() {
		close(outputChan)
		close(errorChan)
	}()
	finished := false
	request.Header.Set("Authorization", "Bearer "+r.ApiKey)
	request.Header.Set("Content-Type", "application/json")
	for !finished {
		response, err := client.Do(request)
		if err != nil {
			fmt.Println("error doing request")
			errorChan <- err
			return
		}
		defer response.Body.Close()
		predictionOutput := new(Output)
		err = json.NewDecoder(response.Body).Decode(&predictionOutput)
		if err != nil {
			fmt.Println("error decoding")
			errorChan <- err
			return
		}
		if finished = r.CheckStatusFinished(predictionOutput); !finished {
			response.Body.Close()
			continue
		}
		outputStr, err := r.ExtractData(predictionOutput)
		if err != nil {
			fmt.Println("Error extracting datap")
			errorChan <- err
			return
		}
		outputChan <- outputStr
	}
}

type predictionRequest[Input any] struct {
	Version string `json:"version,omitempty"`
	Stream  bool   `json:"stream,omitempty"`
	Input   Input  `json:"input"`
}

type predictionResponse[Input any] struct {
	Id        string `json:"id"`
	Model     string `json:"model"`
	Version   string `json:"version"`
	Input     Input  `json:"input"`
	Output    string `json:"output"`
	Logs      string `json:"logs"`
	Error     string `json:"error"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
	Urls      struct {
		Cancel string `json:"cancel"`
		Get    string `json:"get"`
		Stream string `json:"stream"`
	} `json:"urls"`
}

// The default output of the replicate api, I decided to use generics in case a model uses a different output.
type DefaultOutput[Input any] struct {
	Completed_at string `json:"completed_at"`
	Created_at   string `json:"created_at"`
	Data_removed bool   `json:"data_removed"`
	Error        string `json:"error"`
	Id           string `json:"id"`
	Input        Input  `json:"input"`
	Logs         string `json:"logs"`
	Metrics      struct {
		Predict_time float64 `json:"predict_time"`
		Total_time   float64 `json:"total_time"`
	} `json:"metrics"`
	Output     []string `json:"output"`
	Started_at string   `json:"started_at"`
	Status     string   `json:"status"`
	Urls       struct {
		Get    string `json:"get"`
		Cancel string `json:"cancel"`
	} `json:"urls"`
	Version string `json:"version"`
}

// The GetOutput function for the default output struct
func DefaultGetOutput[Input any](output *DefaultOutput[Input]) (string, error) {
	if output.Error != "" {
		return "", errors.New(output.Error)
	}
	return strings.Join(output.Output, ""), nil
}

// The CheckStatusFinished function for the default output struct
func DefaultCheckStatusFinished[Input any](output *DefaultOutput[Input]) bool {
	return output.Status == "succeeded"
}