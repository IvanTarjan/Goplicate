package models

import goplicate "github.com/IvanTarjan/Goplicate"

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

func NewDefaultRequest(prompt, sysprompt string) Llama2_13bInput {
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

func NewLlama2_7bConsumer(apiKey string) *goplicate.ReplicateStreamConsumer[Llama2_13bInput]{
	return goplicate.NewReplicateStreamConsumer(
		apiKey,
		"https://api.replicate.com/v1/models/meta/llama-2-13b-chat/predictions",
		"",
		NewDefaultRequest("Explain to me how to use the llama2 13b model to the best of my knowledge", "You are a helpful, respectful and honest assistant. Always answer as helpfully as possible, while being safe. Your answers should not include any harmful, unethical, racist, sexist, toxic, dangerous, or illegal content. Please ensure that your responses are socially unbiased and positive in nature.\\n\\nIf a question does not make any sense, or is not factually coherent, explain why instead of answering something not correct. If you don't know the answer to a question, please don't share false information."),
		"done",
		"output",
		"error",
	)
}