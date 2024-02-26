![gophish logo](https://raw.github.com/gophish/gophish/master/static/images/gophish_purple.png)

Gophish Docker Usage
=======

# Quickstart

## Local Image Build
```bash
# From project root
docker build -t n4-gophish:latest .
```

## Local Image Run (no data persistence):
```bash
# Navigate to the project root directory
docker run -d --name gophish -p 3333:3333 -p 8080:80 --restart=always n4-gophish:latest
```

## Local Image Run (data persistence):
When running the local image and aiming for data persistence for the first time setup, follow these steps:

**Note**: Creating the local-data/gophish.db file may overwrite existing data if it's already present.
```bash
# Navigate to the project root directory
touch local-data/gophish.db
```

Subsequently, you can run the container with data persistence using the following command:

```bash
# Navigate to the project root directory
docker run -d --name gophish -p 3333:3333 -p 8080:80 --restart=always \
 --mount type=bind,source=./local-data/gophish.db,target=/opt/gophish/gophish.db \
 n4-gophish:latest
```