package goplicate

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

type ReplicateStreamConsumer[Input any] struct {
	ApiKey                                         string
	CreatePredictionUrl                            string
	Version                                        string
	PredictionRequest                              *predictionRequest[Input]
	DoneEventName, OutputEventName, ErrorEventName string
}

func NewReplicateStreamConsumer[Input any](apiKey, createPredictionUrl, version string, baseInput Input, doneEventName, outputEventName, errorEventName string) *ReplicateStreamConsumer[Input] {
	return &ReplicateStreamConsumer[Input]{
		ApiKey:              apiKey,
		CreatePredictionUrl: createPredictionUrl,
		Version:             version,
		PredictionRequest: &predictionRequest[Input]{
			Version: version,
			Stream:  true,
			Input:   baseInput,
		},
		DoneEventName:   doneEventName,
		OutputEventName: outputEventName,
		ErrorEventName:  errorEventName,
	}
}

func (r *ReplicateStreamConsumer[Input]) CreatePrediction(customizeInput func(input *Input)) (*predictionResponse[Input], error) {
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

// func (r *ReplicateStreamConsumer[Input]) CreatePredictionAndGetOutput(customizeInput func(input *Input), outputChan chan string, errorChan chan error) {
// 	predictionUrl, err := r.CreatePrediction(customizeInput)
// 	if err != nil {
// }

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

func doubleNewLineSseScanner(scanner *bufio.Scanner, dataChan chan string, errorChan chan error, doneEventName, outputEventName, errorEventName string) {
	scanner.Split(ScanDoubleNewLine)
	for scanner.Scan() {
		lines := strings.Split(scanner.Text()[1:], "\n")
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

func dropCR(data []byte) []byte {
	if len(data) > 0 && data[len(data)-1] == '\r' {
		return data[0 : len(data)-1]
	}
	return data
}

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
