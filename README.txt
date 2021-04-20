Setup
1. Copy GitHub repository
2. Create .env file in the code folder
3. Add SECRET (used for signing and encrypting JWT - should be of length 32) to .env file
4. Add DBURL to .env file
5. Add APIKEY (from Spoonacular.com) to .env file

Ex. 
.env file format
SECRET=somevalue
DBURL=somevalue
APIKEY=somevalue

To run with Go:
- Run `./main`  from inside the code folder, or build new executable with `go build main` inside the code folder
- Navigate to localhost:8080 (or alternatively, 127.0.0.1:8080)

To run with Docker: 
- Run `docker build -t webserver` from inside the code folder
- Run `docker run -p 8080:8080 -t webserver`
- Navigate to localhost:8080 (or alternatively, 127.0.0.1:8080)

Alternatively, this project can be viewed on the web at: 
https://ChefCart.herokuapp.com

FILES:
code/archive folder -> JS files that are compiled into the templates folder
code/.env -> holds environment variables
code/.gitignore -> list of files to ignore for git
code/Dockerfile -> setup for building docker container
code/go.mod & code/go.sum -> contains information regarding package dependencies
code/main -> executable file to run
code/main.go -> go file to run project
code/webapp/webserver.go -> code holding all webserver operation
code/webapp/templates -> contains HTML templates and JS files for displaying content dynamically
code/webapp/service/jwt.go -> JWT setup and validation
code/webapp/service/loginservice.go -> interface for logging in user
code/webapp/protobuf -> original proto file and auto-generated go protobuf files
code/webapp/models/users.go -> ORM structs for interfacing with the PSQL database
code/webapp/middleware/jwtmid.go -> used to encrypt & decrypt the JWT token
code/webapp/middleware/grpcWeb.go -> used to wrap the grpc server with Gin web server
code/webapp/controller/login.go -> used for extracting login information from form data

All files in the unit_testing folder are stripped down versions of the ones found in the code folder with the exception of:
unit_testing/main_test.go -> contains unit tests for each component of the website