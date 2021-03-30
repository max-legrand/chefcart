Setup:
1. Copy GitHub repository
2. Create .env file in unit_testing folder
3. Add SECRET (used for signing and encrypting JWT - should be of length 32) to .env file
4. Add DBURL to .env file
5. Add APIKEY (from Spoonacular.com) to .env file

To run unit tests: run `go test` from the unit_testing directory