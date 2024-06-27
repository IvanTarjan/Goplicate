package goplicate

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type ReplicateConsumer[Input, Output any] struct {
	ApiKey              string
	CreatePredictionUrl string
	Version             string
	Stream              bool
	PredictionRequest   *predictionRequest[Input]
	CheckStatusFinished         func(output *Output) bool
	ExtractData func(output *Output) (string, error)
}

func NewReplicateConsumer[Input, Output any](apiKey, createPredictionUrl, version string,baseInput Input, extractData func(output *Output) (string, error), checkStatusFinished func(output *Output) bool) *ReplicateConsumer[Input, Output] {
	return &ReplicateConsumer[Input, Output]{
		ApiKey:              apiKey,
		CreatePredictionUrl: createPredictionUrl,
		Version:             version,
		Stream:              false,
		PredictionRequest:   &predictionRequest[Input]{
			Version: version,
			Stream:  false,
			Input:   baseInput,
		},
		ExtractData: extractData,
		CheckStatusFinished: checkStatusFinished,
	}
}

func (r *ReplicateConsumer[Input, Output]) CreatePrediction(customizeInput func(input *Input)) (*predictionResponse[Input], error) {
	customizeInput(&r.PredictionRequest.Input)
	client := &http.Client{}
	requestBodyJson, err := json.Marshal(r.PredictionRequest)
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

func (r *ReplicateConsumer[Input, Output]) GetOutput(outputUrl string, outputChan chan string, errorChan chan error) {
	client := http.Client{}
	request, err := http.NewRequest("GET", outputUrl, nil)
	if err != nil {
		fmt.Println("error creating request")
		errorChan <- err
		return
	}
	defer func ()  {
		close(outputChan)
		close(errorChan)
	}()
	finished := false
	request.Header.Set("Authorization", "Bearer "+r.ApiKey)
	request.Header.Set("Content-Type", "application/json")
	for !finished{
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
