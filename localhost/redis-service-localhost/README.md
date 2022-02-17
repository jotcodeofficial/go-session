// default conf file
https://raw.githubusercontent.com/redis/redis/6.0/redis.conf

//show all images
docker images

// show all running containers
docker ps

// delete old docker images
docker rmi {id or name} -f

// build docker image
docker build  -f Dockerfile -t redis-localhost .

//run the image
docker container run   --publish 6379:6379 -d redis-localhost



//exec inside
docker exec -it <docker-container-nickname> redis-cli

// set value
set name "tester"

// get value
get name

// go into the container to use the redis cli
docker exec -it redis-localhost redis-cli

// kill the process if you want to re-build and re-run
docker kill {id}

// see if any process is running here after killing the docker container
lsof -i :27017

// kill process on machine
kill -9 {pid}
