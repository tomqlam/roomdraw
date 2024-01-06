from sqlalchemy import create_engine
from sqlalchemy import Table, Column, Integer, String, MetaData, ForeignKey
from sqlalchemy import inspect
from sqlalchemy.sql import text

# import env variables
import os
from pathlib import Path

# get env variable called neon_pass from env file
cloud_sql_pass = os.environ.get('CLOUD_SQL_PASS')
cloud_sql_ip = os.environ.get('CLOUD_SQL_IP')

CONNSTR = f'postgresql://postgres:{cloud_sql_pass}@{cloud_sql_ip}/test-db'

engine = create_engine(CONNSTR)

with engine.connect() as connection:
    with open('CreateSuitesTable.sql', 'r') as file:
        query = file.read()
        result = connection.execute(text(query))
        connection.commit()

    with open('CreateGroupsTable.sql', 'r') as file:
        query = file.read()
        result = connection.execute(text(query))
        connection.commit()

    with open('CreateUserTable.sql', 'r') as file:
        query = file.read()
        result = connection.execute(text(query))
        connection.commit()

    with open('CreateRoomTable.sql', 'r') as file:
        query = file.read()
        result = connection.execute(text(query))
        connection.commit()

    