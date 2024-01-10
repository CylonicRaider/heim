FROM postgres:16.1

ENV POSTGRES_PASSWORD heimlich

RUN mkdir -p /docker-entrypoint-initdb.d
ADD create-heim.sql /docker-entrypoint-initdb.d/create-heim.sql
