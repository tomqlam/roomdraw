from sqlalchemy import create_engine
from sqlalchemy import Table, Column, Integer, String, MetaData, ForeignKey
from sqlalchemy import inspect
from sqlalchemy.sql import text

import os

cloud_sql_pass = os.environ.get('CLOUD_SQL_PASS')
cloud_sql_ip = os.environ.get('CLOUD_SQL_IP')

CONNSTR = f'postgresql://postgres:{cloud_sql_pass}@{cloud_sql_ip}/test-db'

engine = create_engine(CONNSTR)
rooms = []
with engine.connect() as connection:
    query = "SELECT room_uuid, suite_uuid FROM Rooms;"
    result = connection.execute(text(query))
    rooms = result.fetchall()
    connection.commit()
# print(rooms)
suite_dict = {}
for room in rooms:
    if room[1] in suite_dict:
        suite_dict[room[1]].append(room[0])
    else:
        suite_dict[room[1]] = [room[0]]
print(suite_dict)

with engine.connect() as connection:
    for suite in suite_dict:
        query = f"UPDATE Suites SET rooms = ARRAY{suite_dict[suite]} WHERE suite_uuid = '{suite}';"
        result = connection.execute(text(query))
        connection.commit()

