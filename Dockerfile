FROM mysql:latest

RUN chown -R mysql:root /var/lib/mysql

ENV MYSQL_DATABASE books
ENV MYSQL_USER mike
ENV MYSQL_PASSWORD mikepass1
ENV MYSQL_ROOT_PASSWORD rootpass

ADD books.sql /etc/mysql/books.sql

RUN cp /etc/mysql/books.sql /docker-entrypoint-initdb.d

EXPOSE 3306
