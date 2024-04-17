docker pull remotepixel/memcached-sasl
docker run --name prom-memcached -p 11211:11211 -e MEMCACHED_USERNAME=dev -e MEMCACHED_PASSWORD=dev -d remotepixel/memcached-sasl memcached -m 512 || docker start prom-memcached
