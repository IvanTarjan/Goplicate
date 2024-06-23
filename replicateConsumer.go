package goplicate

import (
	"bufio"
	"bytes"
	"encoding/json"
	"net/http"
)

type ReplicateConsumer[Input, Output any] struct {
	ApiKey              string
	CreatePredictionUrl string
	Version             string
	Stream              bool
	PredictionRequest   *predictionRequest[Input]
	PredictionResponse  *predictionResponse[Input]
	ExtractDataFromOutput func(output *Output) (string, error)
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
		errorChan <- err
		return
	}
	request.Header.Set("Authorization", "Bearer "+r.ApiKey)
	request.Header.Set("Content-Type", "application/json")
	response, err := client.Do(request)
	if err != nil {
		errorChan <- err
		return
	}
	defer response.Body.Close()
	predictionOutput := new(Output)
	err = json.NewDecoder(response.Body).Decode(&predictionOutput)
	if err != nil {
		errorChan <- err
		return
	}
	outputStr, err := r.ExtractDataFromOutput(predictionOutput)
	if err != nil {
		errorChan <- err
		return
	}
	outputChan <- outputStr
}

func (r *ReplicateConsumer[Input, Output]) GetStreamOutput(outputUrl string, outputChan chan string, errorChan chan error, doneEventName, outputEventName, errorEventName string) {
	request, err := http.NewRequest("GET", outputUrl, nil)
	if err != nil {
		errorChan <- err
		return
	}
	request.Header.Set("Authorization", "Bearer "+r.ApiKey)
	request.Header.Set("Content-Type", "text/event-stream")
	request.Header.Set("Cache-Control", "no-store")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		errorChan <- err
		return
	}
	defer func() {
		response.Body.Close()
		close(outputChan)
		close(errorChan)
	}()

	scanner := bufio.NewScanner(response.Body)
	doubleNewLineSseScanner(scanner, outputChan, errorChan, doneEventName, outputEventName, errorEventName)
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
