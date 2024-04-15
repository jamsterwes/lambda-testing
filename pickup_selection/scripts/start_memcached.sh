docker pull memcached
docker run --name prom-memcached -p 11211:11211 -d memcached memcached -m 512 || docker start prom-memcached
