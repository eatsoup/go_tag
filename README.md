# GO TAG
## Simple tool to tag a remote docker image without pulling it first

## How to use
### Tool
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
