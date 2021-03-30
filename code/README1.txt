ChefCart

Setup
1. Copy GitHub repository
2. Create .env file in code folder
3. Add SECRET (used for signing and encrypting JWT - should be of length 32) to .env file
4. Add DBURL to .env file
5. Add APIKEY (from Spoonacular.com) to .env file

To run ChefCart with Go
Run `./main`  from inside the code folder, or build new executable with `go build main`
Navigate to localhost:8080 (or alternatively, 127.0.0.1:8080)

To run ChefCart with Docker
Run `docker build -t webserver` from inside the code folder
Run `docker run -p 8080:8080 -t webserver`
Navigate to localhost:8080 (or alternatively, 127.0.0.1:8080)
