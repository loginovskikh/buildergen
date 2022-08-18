package main

import (
	"flag"
	"fmt"
	"log"

	generatorPkg "gitlab.ozon.dev/loginovskikh/buildergen/generator"
)

func main() {
	sourcePtr := flag.String("source", "", "path to source file with type definition")
	typePtr := flag.String("type", "", "name of custom type in source file")
	outputFileName := flag.String("o", "", "path to file with output")
	helpCmd := flag.Bool("help", false, "help message")

	flag.Parse()

	if *helpCmd {
		fmt.Println(`
Buider generator is util thats hepls devs to perform routine tasks 
like writing struct types builders.
	
Available flags:
-source - path to source file (*required)    
-type - name of the type to generate builder for (*required)    
-o - name of file, where builder will be stored. Default value is "builder.go"    
-help - show help message    

buildergen -source ./core/user/user.go -type User -o user_builder.go

This example command will generate bilder for type User which is defined in file ./core/user/user.go and save builder
in file user_builder.go in source file folder`)
	} else {

		if *sourcePtr == "" {
			log.Fatal("empty [source] argument. use -source path/to/file.go")
		}

		if *typePtr == "" {
			log.Fatal("empty [type] argument. use -type User")
		}

		if *outputFileName == "" {
			*outputFileName = "builder.go"
		}

		generator := generatorPkg.NewGenerator(*sourcePtr, *typePtr, *outputFileName)

		err := generator.Generate()
		if err != nil {
			log.Fatal("error while generating builder: ", err.Error())
		}
	}
}
