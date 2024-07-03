package llama213b

import (
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

// NewLlama2_13bDefaultRequest creates a request with the default values except from the prompt and sysprompt to ease the use of the package.
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

func NewLlama2_13bConsumer(apiKey string, input Llama2_13bInput) *goplicate.ReplicateConsumer[Llama2_13bInput, goplicate.DefaultOutput[Llama2_13bInput]] {
	return goplicate.NewReplicateConsumer(
		apiKey,
		"https://api.replicate.com/v1/models/meta/llama-2-13b-chat/predictions",
		"",
		input,
		goplicate.DefaultGetOutput[Llama2_13bInput],
		goplicate.DefaultCheckStatusFinished[Llama2_13bInput],
	)
}
