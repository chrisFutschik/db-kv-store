Problem Assumptions:
1. Value has no minimum requirement, can submit key with empty value
2. User will input values, these will not be randomly generated
3. No constraints about updating previous keys, exsisting keys will be updated when PUT command is used. 

Running:
> Server runs on "localhost:3000"
> "./kvBackUp.json" is key-value file store

Required packages:
> 1. github.com/gin-gonic/gin

To run server: 
> go run main.go

Commands: 

1. Write/Update a value to the key foo:
$ curl http://localhost:3000 --include --header "Content-Type: application/json" --request "POST" --data '{"key": "foo","value": "Hello World"}'

2. Fetch the value of foo:
$ curl -i http://localhost:3000/foo

4. Delete the key foo:
$ curl -i -X DELETE http://localhost:3000/foo