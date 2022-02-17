// kill the process if you want to re-build and re-run
docker kill {id}

// see if any process is running here after killing the docker container
lsof -i :27017

// kill process on machine
kill -9 {pid}


------------------------------------------------
DOCKER SHORTCUTS

//show all images
docker images

// show all running containers
docker ps

// delete old docker images
docker rmi {id or name} -f

// build docker image
docker build  -f Dockerfile -t mongo-localhost .

//run the image
docker container run --env-file mongo-localhost-auth.env --publish 27017:27017 -d mongo-localhost
