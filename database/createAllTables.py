from dotenv import load_dotenv
from sqlalchemy import create_engine
from sqlalchemy.sql import text

# import env variables
import os
from pathlib import Path

# load env variables
env_path = Path('.') / '.env'
load_dotenv(dotenv_path=env_path, verbose=True)

sql_pass = os.environ.get('SQL_PASS')
sql_ip = os.environ.get('SQL_IP')
sql_db_name = os.environ.get('SQL_DB_NAME')
sql_user = os.environ.get('SQL_USER')

CONNSTR = f'postgresql://{sql_user}:{sql_pass}@{sql_ip}/{sql_db_name}'

print(CONNSTR)
engine = create_engine(CONNSTR)

with engine.connect() as connection:
    with open('DropTables.sql', 'r', encoding='utf-8') as file:
        query = file.read()
        result = connection.execute(text(query))
        connection.commit()

    with open('CreateSuitesTable.sql', 'r', encoding='utf-8') as file:
        query = file.read()
        result = connection.execute(text(query))
        connection.commit()

    with open('CreateGroupsTable.sql', 'r', encoding='utf-8') as file:
        query = file.read()
        result = connection.execute(text(query))
        connection.commit()

    with open('CreateUserTable.sql', 'r', encoding='utf-8') as file:
        query = file.read()
        result = connection.execute(text(query))
        connection.commit()

    with open('CreateRoomTable.sql', 'r', encoding='utf-8') as file:
        query = file.read()
        result = connection.execute(text(query))
        connection.commit()

    with open('CreateRateLimitTable.sql', 'r', encoding='utf-8') as file:
        query = file.read()
        result = connection.execute(text(query))
        connection.commit() 
