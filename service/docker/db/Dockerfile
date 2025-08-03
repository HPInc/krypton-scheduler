FROM bitnami/cassandra:4.0

ENV LANG C.UTF-8

# Here, you can add any *.sh or *.cql scripts inside /docker-entrypoint-initdb.d
#  *.sh files will be executed BEFORE launching cassandra
#  *.cql files will be executed with cqlsh -f AFTER cassandra started
# Files are executed in name order (ls * | sort)
COPY *.cql /docker-entrypoint-initdb.d/
