# goplicate

goplicate is a Go package designed to interact with the Replicate API, enabling users to create predictions and retrieve outputs. The package leverages Go generics to provide a flexible interface for different input and output types, making it adaptable to various use cases.

## Features

- Generic Consumer: Use the ReplicateConsumer struct to create and manage API requests with customizable input and output types.
- Default Implementations: Includes default functions for extracting data and checking prediction status.
- Asynchronous Output Retrieval: Fetch prediction outputs asynchronously using goroutines.

## Installation

To install the goplicate package, use:

  go get github.com/IvanTarjan/Goplicate
