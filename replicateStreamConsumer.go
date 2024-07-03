package goplicate

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

// While this generic struct can be used by itself, it is recommended to make a package with a concrete implementation.
//
// The replicateConsumer struct is a Generic consumer that receives makes requests to the replicate api using the Input and Output types provided.
//
// The version attribute is optional, if it is not needed leave it as an empty string.
//
// The event names are required to identify them and send them to their corresponding channels. 
type ReplicateStreamConsumer[Input any] struct {
	ApiKey                                         string
	CreatePredictionUrl                            string
	Version                                        string
	BasePredictionRequest                              *predictionRequest[Input]
	DoneEventName, OutputEventName, ErrorEventName string
}
// While this generic struct can be used by itself, it is recommended to make a package with a concrete implementation.
//
// The replicateConsumer struct is a Generic consumer that receives makes requests to the replicate api using the Input and Output types provided.
//
// The version attribute is optional, if it is not needed leave it as an empty string.
//
// The event names are required to identify them and send them to their corresponding channels. 
func NewReplicateStreamConsumer[Input any](apiKey, createPredictionUrl, version string, baseInput Input, doneEventName, outputEventName, errorEventName string) *ReplicateStreamConsumer[Input] {
	return &ReplicateStreamConsumer[Input]{
		ApiKey:              apiKey,
		CreatePredictionUrl: createPredictionUrl,
		Version:             version,
		BasePredictionRequest: &predictionRequest[Input]{
			Version: version,
			Stream:  true,
			Input:   baseInput,
		},
		DoneEventName:   doneEventName,
		OutputEventName: outputEventName,
		ErrorEventName:  errorEventName,
	}
}

// Create prediction receives a function that allows you to customize only the desired fields of the input without changing the rest of the prediction request and input.
//
// The function returns a prediction response that contains the urls needed to get the output of the prediction.
func (r *ReplicateStreamConsumer[Input]) CreatePrediction(customizeInput func(input Input) Input) (*predictionResponse[Input], error) {
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

//GetStreamOutput receives the url of the output and channels to send the output and errors
//
//The function channels every bit of data of the events sent by the api as soon as they arribe to let you return the stream the same way that you receive it
func (r *ReplicateStreamConsumer[Input]) GetStreamOutput(outputUrl string, outputChan chan string, errorChan chan error) {
	request, err := http.NewRequest("GET", outputUrl, nil)
	if err != nil {
		errorChan <- err
		return
	}
	request.Header.Set("Authorization", "Bearer "+r.ApiKey)
	request.Header.Set("Accept", "text/event-stream")
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
	doubleNewLineSseScanner(scanner, outputChan, errorChan, r.DoneEventName, r.OutputEventName, r.ErrorEventName)
}

type serverSentEvent struct {
	Event string `json:"event"`
	Data string `json:"data"`
}
// The function separated the flow of data into different events and routes them to their corresponding channels.
func doubleNewLineSseScanner(scanner *bufio.Scanner, dataChan chan string, errorChan chan error, doneEventName, outputEventName, errorEventName string) {
	scanner.Split(ScanDoubleNewLine)
	for scanner.Scan() {
		lines := strings.Split(string(scanner.Bytes()[1:]), "\n")
		event := serverSentEvent{}
		for _, line := range lines {
			if strings.HasPrefix(line, "event: ") {
				event.Event = line[7:]
			}
			if strings.HasPrefix(line, "data: ") {
				event.Data += line[6:]
			}
		}
		switch event.Event {
		case doneEventName:
			return
		case outputEventName:
			dataChan <- event.Data
		case errorEventName:
			errorChan <- errors.New(event.Data)
			return
		}
	}
}

// dropCR drops a terminal \r from the data. 
func dropCR(data []byte) []byte {
	if len(data) > 0 && data[len(data)-1] == '\r' {
		return data[0 : len(data)-1]
	}
	return data
}

// ScanDoubleNewLine is a bufio.SplitFunc that splits on a double newline, which is required to convert the flow of data into separate events
func ScanDoubleNewLine(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.Index(data, []byte("\n\n")); i >= 0 {
		// We have a full newline-terminated line.
		return i + 1, dropCR(data[0:i]), nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), dropCR(data), nil
	}
	// Request more data.
	return 0, nil, nil
}
