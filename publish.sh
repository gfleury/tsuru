COMMIT=$(git log --pretty=format:"%h" -1)
REPO=gfleury/tsuru-api
TAG="latest"
docker build -f Dockerfile.dev -t $REPO:$COMMIT . 
docker tag $REPO:$COMMIT $REPO:$TAG
docker push $REPO

