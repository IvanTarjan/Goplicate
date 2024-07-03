# goplicate

goplicate is a Go package designed to interact with the Replicate API, enabling users to create predictions and retrieve outputs. The package leverages Go generics to provide a flexible interface for different input and output types, making it adaptable to various use cases.

## Features

- Generic Consumer: Use the ReplicateConsumer struct to create and manage API requests with customizable input and output types.
- Default Implementations: Includes default functions for extracting data and checking prediction status.
- Asynchronous Output Retrieval: Fetch prediction outputs asynchronously using goroutines.

## Installation

To install the goplicate package, use:

    go get github.com/IvanTarjan/Goplicate

## Usage

Usage example making requests to the llama-2-13b model:

    input := llama213b.NewLlama2_13bDefaultRequest("Explain to me how to use the llama2 13b model to the best of my knowledge <Your prompt>", "You are a helpful, respectful and honest assistant. <Your sysprompt>")
	llama2Consumer := llama213b.NewLlama2_13bConsumer("<Your Api Key>", input)
	predictionResponse, err := llama2Consumer.CreatePrediction(func(input llama213b.Llama2_13bInput) llama213b.Llama2_13bInput { return input })
	if err != nil {
		panic(err)
	}
	data2Chan := make(chan string)
	err2Chan := make(chan error)
	go llama2Consumer.GetOutput(predictionResponse.Urls.Get, data2Chan, err2Chan)

	select {
	case data := <-data2Chan:
		fmt.Print(data)
	case err := <-err2Chan:
		fmt.Println(err)
		return
	}

Usage example making stream requests to the llama-2-13b model:

 	input := llama213b.NewLlama2_13bDefaultRequest("Explain to me how to use the llama2 13b model to the best of my knowledge <Your prompt>", "You are a helpful, respectful and honest assistant. (<Your sysprompt>")
	llama2StreamConsumer := llama213b.NewLlama2_13bStreamConsumer("<Your Api Key>", input)
	predictionStreamResponse, err := llama2StreamConsumer.CreatePrediction(func(input llama213b.Llama2_13bInput) llama213b.Llama2_13bInput {return input})
	if err != nil {
		panic(err)
	}

	dataChan := make(chan string)
	errChan := make(chan error)
	go llama2StreamConsumer.GetStreamOutput(predictionStreamResponse.Urls.Stream, dataChan, errChan)

	isRunning := true
	for isRunning{
		select {
		case data, ok := <-dataChan:
			if !ok {
				isRunning = false
				break
			}
			fmt.Print(data)
		case err, ok := <-errChan:
			if !ok {
				isRunning = false
				break
			}
			fmt.Println(err)
			return
		}
	}

 ## Implementing other models

 Implementation example of the stable-diffusion-3 model

    package stablediffusion3
    
    import (
    	goplicate "github.com/IvanTarjan/Goplicate"
    )
    
    // Make a struct of the input json schema
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
    
    // If you want to (But it is really recommended) you can make a function that returns an input struct with the model's default values
    func NewStableDiffusion3DefaultRequest(prompt string) StableDiffusion3Input {
    	return StableDiffusion3Input{
    		Cfg:            4.5,
    		Seed:           448433150,
    		Image:          "",
    		Steps:          28,
    		Prompt:         prompt,
    		AspectRatio:    "3:2",
    		OutputFormat:   "webp",
    		OutputQuality:  79,
    		NegativePrompt: "ugly, distorted, low details",
    		PromptStrength: 0.85,
    	}
    }

    // Lastly, making a function to return the consumer using the input and output structs of the model, In this case the model returned the default output schema so I used the default get output and check status finished
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


### Using the newly Implemented model

    sdInput := stablediffusion3.NewStableDiffusion3DefaultRequest("An artwork of a dwarf cooking monsters, ukiyo-e style")
	stableDiffusionConsumer := stablediffusion3.NewStableDiffusion3Consumer("<Yout Api Key>", sdInput)
	predictionResponse, err := stableDiffusionConsumer.CreatePrediction(func(input stablediffusion3.StableDiffusion3Input) stablediffusion3.StableDiffusion3Input {
		input.Seed = rand.Int()
		return input
	})
	if err != nil {
		panic(err)
	}
	sdDataChan := make(chan string)
	sdErrChan := make(chan error)
	go stableDiffusionConsumer.GetOutput(predictionResponse.Urls.Get, sdDataChan, sdErrChan)
	select{
		case data:= <-sdDataChan:
			fmt.Println(data)
		case err:= <-sdErrChan:
			fmt.Println(err)
	}
