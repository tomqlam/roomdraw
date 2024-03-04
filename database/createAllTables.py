from dotenv import load_dotenv
from sqlalchemy import create_engine
from sqlalchemy import Table, Column, Integer, String, MetaData, ForeignKey
from sqlalchemy import inspect
from sqlalchemy.sql import text

# import env variables
import os
from pathlib import Path

# load env variables
env_path = Path('.') / '.env'
load_dotenv(dotenv_path=env_path, verbose=True)

cloud_sql_pass = os.environ.get('CLOUD_SQL_PASS')
cloud_sql_ip = os.environ.get('CLOUD_SQL_IP')
cloud_sql_db_name = os.environ.get('CLOUD_SQL_DB_NAME')
cloud_sql_user = os.environ.get('CLOUD_SQL_USER')

CONNSTR = f'postgresql://{cloud_sql_user}:{cloud_sql_pass}@{cloud_sql_ip}/{cloud_sql_db_name}'

print(CONNSTR)
engine = create_engine(CONNSTR)

with engine.connect() as connection:
    with open('DropTables.sql', 'r') as file:
        query = file.read()
        result = connection.execute(text(query))
        connection.commit()
        
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

    