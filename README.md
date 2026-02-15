# Go DTO Generator (dtogen)

`dtogen` is a CLI utility to generate DTOs and mapper functions from Go structs.

## Features
*   **AST-based Parsing**: Uses `go/packages` to parse source code without requiring it to be compiled into the tool.
*   **Customizable Output**: Supports renaming the output struct and file.
*   **Field Exclusion**: Allows excluding specific fields (e.g., sensitive data like passwords).
*   **Field Inclusion**: Allows specifying an allowlist of fields to include.
*   **Field Renaming**: Supports renaming fields in the generated DTO.
*   **YAML Configuration**: Supports batch generation via a YAML configuration file.
*   **Template Support**: Uses a default embedded template but supports custom templates via the `-template` flag.

## Usage

### Build
```bash
go build -o dtogen .
```

### Run
```bash
# Long flags
./dtogen --src ./path/to/models --type User --exclude Password --out user_dto.go

# Short flags
./dtogen -s ./path/to/models -t User -e Password -o user_dto.go
# Output will be in ./path/to/models/user_dto.go by default

# With output directory
./dtogen -s ./path/to/models -t User -d ./generated -o user_dto.go

# Include specific fields and rename
./dtogen -s ./path/to/models -t User -i ID,Username -r Username:LoginName -o user_dto.go

# Generate sample config
./dtogen --init

# Run with config file
./dtogen --config dtogen.yaml
```

### Example Output
Given a `User` struct:
```go
type User struct {
    ID       int
    Username string
    Password string
}
```

Running `dtogen -t User -e Password` generates:
```go
type UserDTO struct {
    ID       int
    Username string
}

func ToUserDTO(v *User) *UserDTO {
    if v == nil {
        return nil
    }
    return &UserDTO{
        ID:       v.ID,
        Username: v.Username,
    }
}
```
