# GO TAG
## Simple tool to tag a remote docker image without pulling it first
## Easy way to speed up your CI/CD!

## How to use
This tool can be used for the Docker Hub, authenticated private registries and unauthenticated registries.
### Docker hub
```bash
export REGISTRY_USER=your_docker.io_username
export REGISTRY_PASSWORD=your_docker.io_password
./go_tag your_repo/your_image:old_tag your_repo/your_image:new_tag
```
### Private authenicated registry
```bash
export REGISTRY=https://your_registry_url
export REGISTRY_USER=your_registry_username
export REGISTRY_PASSWORD=your_registry_password
./go_tag your_repo/your_image:old_tag your_repo/your_image:new_tag
```
### Private unauthenicated registry
```bash
export REGISTRY=https://your_registry_url
./go_tag your_repo/your_image:old_tag your_repo/your_image:new_tag
```

## How to run / install
### From source
```bash
git clone https://github.com/eatsoup/go_tag.git
cd go_tag && go build .
export REGISTRY_USER='username'
export REGISTRY_PASSWORD='password'
./go_tag repository/image:oldtag repository_image:newtag
```
### Docker image (6MB)
```bash
docker run --rm -t -e REGISTRY_USER='username' -e REGISTRY_PASSWORD='password' eatsoup/go_tag go_tag repository/image:old_tag repository/image:new_tag
```
