FROM postgres:13

# Dependencies
RUN apt-get update && apt-get install -y \
    build-essential \
    git \
    wget \
    postgresql-server-dev-13 \
    && rm -rf /var/lib/apt/lists/*

# Postgres Docker Images copy contents of postgresql.conf.sample to postgresql.conf during initialization
# COPY custom.conf /usr/share/postgresql/postgresql.conf.sample -- DO NOT USE
RUN echo "shared_preload_libraries = 'pg_stat_statements,pg_wait_sampling,pg_stat_monitor'" >> /usr/share/postgresql/postgresql.conf.sample
RUN echo "pg_stat_statements.track = all" >> /usr/share/postgresql/postgresql.conf.sample
RUN echo "pg_stat_statements.save = on" >> /usr/share/postgresql/postgresql.conf.sample
RUN echo "pg_stat_monitor.pgsm_enable_query_plan = on" >> /usr/share/postgresql/postgresql.conf.sample

# Install pg_wait_sampling
RUN git clone https://github.com/postgrespro/pg_wait_sampling.git \
    && cd pg_wait_sampling \
    && make USE_PGXS=1 \
    && make USE_PGXS=1 install \
    && cd .. \
    && rm -rf pg_wait_sampling

# Install pg_stat_monitor
RUN git clone https://github.com/percona/pg_stat_monitor.git \
    && cd pg_stat_monitor \
    && make USE_PGXS=1 \
    && make USE_PGXS=1 install \
    && cd .. \
    && rm -rf pg_stat_monitor

# Download the titanic database
RUN wget https://raw.githubusercontent.com/neondatabase/postgres-sample-dbs/main/titanic.sql -P /docker-entrypoint-initdb.d/

RUN echo ${PWD} && ls -lR

# Enable the extensions and setup the titanic database
COPY 01-init-extensions.sql /docker-entrypoint-initdb.d/01-init-extensions.sql
COPY 02-create-database.sql /docker-entrypoint-initdb.d/02-create-database.sql
COPY 03-import-data.sql /docker-entrypoint-initdb.d/03-import-data.sql