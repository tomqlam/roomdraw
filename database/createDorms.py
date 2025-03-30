import os
from sqlalchemy import create_engine
from sqlalchemy.sql import text
from uuid import uuid4
from dotenv import load_dotenv

dotenv_path = os.path.join(os.path.dirname(__file__), '.env')
print(dotenv_path)

load_dotenv(dotenv_path=dotenv_path, verbose=True)

sql_pass = os.environ.get('SQL_PASS')
sql_ip = os.environ.get('SQL_IP')
sql_db_name = os.environ.get('SQL_DB_NAME')
sql_user = os.environ.get('SQL_USER')

CONNSTR = f'postgresql://{sql_user}:{sql_pass}@{sql_ip}/{sql_db_name}'

engine = create_engine(CONNSTR)

# List
# 1 = East
# 2 = North
# 3 = South
# 4 = West
# 5 = Atwood
# 6 = Sontag
# 7 = Case
# 8 = Drinkward
# 9 = Linde
# 10 = Garrett House


def populate_using_json(dorm_id: int, dorm_name: str, json_file: str):
    import json
    with open(json_file, 'r', encoding='utf-8') as file:
        data = json.load(file)
        with engine.connect() as connection:
            query = f"DELETE FROM Rooms WHERE dorm = {dorm_id}; DELETE FROM Suites WHERE dorm = {dorm_id};"
            result = connection.execute(text(query))
            # connection.commit()
            for floor in range(len(data["floors"])):
                for suite in data["floors"][floor]["suites"]:
                    # generate a uuid for the suite
                    suite_uuid = uuid4()
                    query = f"INSERT INTO Suites (suite_uuid, dorm, dorm_name, room_count, floor, alternative_pull, can_lock_pull) VALUES ('{suite_uuid}' ,{dorm_id}, '{dorm_name}', {len(suite['rooms'])}, {floor}, {suite['alternative_pull']}, {suite['can_lock_pull']}) RETURNING suite_uuid;"
                    result = connection.execute(text(query))
                    suite_uuid = result.fetchone()[0]
                    # connection.commit()

                    room_uuids = []

                    # insert the rooms into the rooms table
                    for room in suite["rooms"]:
                        # insert the room into the rooms table
                        room_uuid = uuid4()
                        query = f"INSERT INTO Rooms (room_uuid, dorm, dorm_name, room_id, suite_uuid, max_occupancy, current_occupancy, frosh_room_type) VALUES ('{room_uuid}' ,{dorm_id}, '{dorm_name}', '{room['room_number']}', '{suite_uuid}', {room['capacity']}, 0, {room['frosh_room_type']}) RETURNING room_uuid;"
                        room_uuids.append(connection.execute(
                            text(query)).fetchone()[0])
                        # connection.commit()

                    room_uuid_string = ""
                    for room_uuid in room_uuids:
                        room_uuid_string += f"'{room_uuid}'::UUID, "
                    room_uuid_string = room_uuid_string[:-2]
                    # update the suite which contains the room
                    query = f"UPDATE Suites SET rooms = ARRAY[{room_uuid_string}] WHERE suite_uuid = '{suite_uuid}';"
                    connection.execute(text(query))
            connection.commit()


def populate_all():
    populate_using_json(1, "East", "east.json")
    populate_using_json(2, "North", "north.json")
    populate_using_json(3, "South", "south.json")
    populate_using_json(4, "West", "west.json")
    populate_using_json(5, "Atwood", "atwood.json")
    populate_using_json(6, "Sontag", "sontag.json")
    populate_using_json(7, "Case", "case.json")
    populate_using_json(8, "Drinkward", "drinkward.json")
    populate_using_json(9, "Linde", "linde.json")


populate_all()
