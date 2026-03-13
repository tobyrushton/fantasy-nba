source .env

docker build -t fantasy-nba-db .

docker run -d \
    --name fantasy-nba-db \
    -e POSTGRES_USER=admin \
    -e POSTGRES_PASSWORD=$DB_PASSWORD \
    -e POSTGRES_DB=mydb \
    -p 5432:5432 \
    fantasy-nba-db