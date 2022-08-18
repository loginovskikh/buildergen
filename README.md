# Builder generator

Builder generator is util that helps devs to perform routine tasks 
like writing struct types builders.

### Installation

```bash
go install github.com/loginovskikh/buildergen
```

	
### Available flags:   
``` 
-source - path to source file (*required)    
-type - name of the type to generate builder for (*required)    
-o - name of file, where builder will be stored. Default value is "builder.go"    
-help - show help message    
```
### Example:
```bash
buildergen -source ./core/user/user.go -type User -o user_builder.go
```
This example command will generate builder for type ***User*** which is defined in file ***./core/user/user.go*** and save builder
in file ***user_builder.go*** in source file folder


Also command can be used with ***go:generate***
```go
//go:generate buildergen -source ./user.go -type User -o user_builder.go
```

### Limitations

1. Can be used for one type at a time
2. Saving the result is possible only in the source file folder 


