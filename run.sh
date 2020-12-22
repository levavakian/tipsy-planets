if [ "$(docker ps -a | grep tipc)" ]
then
  docker start tipc > /dev/null
  docker exec -it tipc bash
else
  docker run -it --name tipc -e DISPLAY --privileged --user $(id -u):$(id -g) -v $(pwd)/../tipsy-planets:/home/apps/tipsy-planets --network=host tipsy-planets:latest bash
fi