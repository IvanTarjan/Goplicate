package recommendations

import (
	"bufio"
	"fmt"
	"net/http"
	"strings"
)

// ver de adaptar para que reciba la url o solo usar lo de adentro
func readStreamEvents() (string, error) {

	url := "http://stream-del-result-get.com/steream"

	// ver de instanciar cliente o hacer el get de una
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	defer resp.Body.Close()

	// el viejo bufio, nada le gana
	reader := bufio.NewReader(resp.Body)
	var data string
	var completeMessage strings.Builder

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			// aca el reader deberia detectar el eof o un error, lo que pase primero y detiene el for.
			fmt.Println("Error: ", err)
			return "", err
		}

		// por las dudas que hayan espacios innecesarios, ver si hay que quitarlos o no si las palabras quedan pegadas
		line = strings.TrimSpace(line)
		// este if sirve para saber cuando comienza y termina un "bloque" del stream de datos(conjunto de event, id y data).
		if line == "" {
			if data != "" {
				// la data de aca es la que recibis despues de la primera vuelta
				completeMessage.WriteString(data)
			}
			// reseteo la data (me hace acordar a javascript y html)
			data = ""
			// Esta parte es descartable, pero sirve para ver en consola lo que se recibe y se escribe arriba.
			fmt.Printf("Data: %s\n", data)
			continue
		}

		// Aca detectas y parseas la linea data quitandole el prefijo data: y lo guardas en data que ahi veras si returneas data o la usas directamente en algun lado y modificas la firma de la funcion.
		if strings.HasPrefix(line, "data: ") {
			// este checkeo es un por las dudas, no recuerdo bien si algun data era vacio, a veces el primero o si el "hi" del principio contaba para algo, si ves que no es necesario AFUERA.
			if data == "" {
				data = strings.TrimPrefix(line, "data: ")
			} else {
				// quiza aca ver de apendear o agregar algo como un espacio entre cada una, dependiendo de que string se apendea al final
				data += strings.TrimPrefix(line, "data: ")
			}
		}
	}
}
