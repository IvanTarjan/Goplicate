package models

import (
	"errors"
	"strings"

	goplicate "github.com/IvanTarjan/Goplicate"
)

type Llama2_13bInput struct {
	Top_k            int     `json:"top_k"`
	Top_p            int     `json:"top_p"`
	Prompt           string  `json:"prompt"`
	Temperature      float64 `json:"temperature"`
	System_prompt    string  `json:"system_prompt"`
	Length_penalty   int     `json:"length_penalty"`
	Max_new_tokens   int     `json:"max_new_tokens"`
	Prompt_template  string  `json:"prompt_template"`
	Presence_penalty int     `json:"presence_penalty"`
}
type Llama2_13bOutput struct {
	Completed_at string          `json:"completed_at"`
	Created_at   string          `json:"created_at"`
	Data_removed bool            `json:"data_removed"`
	Error        string          `json:"error"`
	Id           string          `json:"id"`
	Input        Llama2_13bInput `json:"input"`
	Logs         string          `json:"logs"`
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

func NewLlama2_13bDefaultRequest(prompt, sysprompt string) Llama2_13bInput {
	return Llama2_13bInput{
		0,
		1,
		prompt,
		0.75,
		sysprompt,
		1,
		500,
		"<s>[INST] <<SYS>>\n{system_prompt}\n<</SYS>>\n\n{prompt} [/INST]",
		0,
	}
}

func NewLlama2_13bStreamConsumer(apiKey string, input Llama2_13bInput) *goplicate.ReplicateStreamConsumer[Llama2_13bInput] {
	return goplicate.NewReplicateStreamConsumer(
		apiKey,
		"https://api.replicate.com/v1/models/meta/llama-2-13b-chat/predictions",
		"",
		input,
		"done",
		"output",
		"error",
	)
}

func NewLlama2_13bConsumer(apiKey string, input Llama2_13bInput) *goplicate.ReplicateConsumer[Llama2_13bInput, Llama2_13bOutput] {
	return goplicate.NewReplicateConsumer(
		apiKey,
		"https://api.replicate.com/v1/models/meta/llama-2-13b-chat/predictions",
		"",
		input,
		func(output *Llama2_13bOutput) (string, error) {
			if output.Error != "" {
				return "", errors.New(output.Error)
			}
			return strings.Join(output.Output, ""), nil
		},
		func(output *Llama2_13bOutput) bool {
			return output.Status == "succeeded"
		},
	)
}
