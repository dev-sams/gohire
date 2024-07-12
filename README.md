# Installation

1. Clone the repo
```bash
git clone https://github.com/dev-sams/gohire.git
```
2. Download the dependencies
```go
go mod download
```

3. Run the server
```go
go run server.go
```
4. Server will be running at `http://localhost:3001`

## Dependecies

#### Make sure you have
`gcc` installed and `cgo` enabled

## Testing
1. Run following command to execute test file.
   ```go
   go test -v
   ```
