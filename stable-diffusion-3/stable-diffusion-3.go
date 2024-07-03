package stablediffusion3

import (
	goplicate "github.com/IvanTarjan/Goplicate"
)

// As this are the only supported output formats I'll leave them as constants to make them easier to use
const (
	OutputFormatJpg  = "jpg"
	OutputFormatWebp = "webp"
	OutputFormatPng  = "png"
)

// As this are the only supported aspect ratios I'll leave them as constants to make them easier to use
const (
	AspectRatio1_1  = "1:1"
	AspectRatio16_9 = "16:9"
	AspectRatio21_9 = "21:9"
	AspectRatio2_3  = "2:3"
	AspectRatio3_2  = "3:2"
	AspectRatio4_5  = "4:5"
	AspectRatio5_4  = "5:4"
	AspectRatio9_16 = "9:16"
	AspectRatio9_21 = "9:21"
)

type StableDiffusion3Input struct {
	Cfg            float64 `json:"cfg"`
	Seed           int     `json:"seed"`
	Image          string  `json:"image,omitempty"`
	Steps          int     `json:"steps"`
	Prompt         string  `json:"prompt"`
	AspectRatio    string  `json:"aspect_ratio"`
	OutputFormat   string  `json:"output_format"`
	OutputQuality  int     `json:"output_quality"`
	NegativePrompt string  `json:"negative_prompt"`
	PromptStrength float64 `json:"prompt_strength"`
}

// NewStableDiffusion3DefaultRequest creates a request with the default values except from the prompt to ease the use of the package.
func NewStableDiffusion3DefaultRequest(prompt string) StableDiffusion3Input {
	return StableDiffusion3Input{
		Cfg:            4.5,
		Seed:           448433150,
		Image:          "",
		Steps:          28,
		Prompt:         prompt,
		AspectRatio:    AspectRatio3_2,
		OutputFormat:   OutputFormatWebp,
		OutputQuality:  79,
		NegativePrompt: "ugly, distorted, low details",
		PromptStrength: 0.85,
	}
}

func NewStableDiffusion3Consumer(apiKey string, input StableDiffusion3Input) *goplicate.ReplicateConsumer[StableDiffusion3Input, goplicate.DefaultOutput[StableDiffusion3Input]] {
	return goplicate.NewReplicateConsumer(
		apiKey,
		"https://api.replicate.com/v1/models/stability-ai/stable-diffusion-3/predictions",
		"",
		input,
		goplicate.DefaultGetOutput[StableDiffusion3Input],
		goplicate.DefaultCheckStatusFinished[StableDiffusion3Input],
	)
}