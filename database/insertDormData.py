from typing import List, Dict
from sqlalchemy import create_engine
from sqlalchemy import Table, Column, Integer, String, MetaData, ForeignKey
from sqlalchemy import inspect
from sqlalchemy.sql import text

# import env variables
import os
from dotenv import load_dotenv
from pathlib import Path

dotenv_path = os.path.join(os.path.dirname(__file__), '.env')
print(dotenv_path)

load_dotenv(dotenv_path=dotenv_path, verbose=True)

cloud_sql_pass = os.environ.get('CLOUD_SQL_PASS')
cloud_sql_ip = os.environ.get('CLOUD_SQL_IP')

CONNSTR = f'postgresql://postgres:{cloud_sql_pass}@{cloud_sql_ip}/test-db'

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
# 10 = Garett House

# All East suites have two rooms, floor just adds 50 to the room number
def create_normal_inner_suite(starting_room_number_without_floor: int, order: List[int], floor: int, room_offset: int, dorm: int, alternative: bool, connection: object):
    # start a transaction
    if dorm not in [1, 2, 4]:
        raise Exception("Invalid dorm")
    if dorm == 1:
        dorm_name = "East"
        floor_increment = 100
    elif dorm == 2:
        dorm_name = "North"
        floor_increment = 200
    elif dorm == 4:
        dorm_name = "West"
        floor_increment = 400
    
    query = f"INSERT INTO Suites (dorm, dorm_name, room_count, floor, alternative_pull) VALUES ({dorm}, '{dorm_name}', {len(order)}, {floor}, {alternative}) RETURNING suite_uuid;"
    result = connection.execute(text(query))
    suite_uuid = result.fetchone()[0]
    # connection.commit()

    room_uuids = []

    # insert the rooms into the rooms table
    actual_starting = starting_room_number_without_floor + floor_increment
    if floor == 1:
        actual_starting += 50
    for room_type in order:
        # get the room number
        room_number = actual_starting
        # insert the room into the rooms table
        query = f"INSERT INTO Rooms (dorm, dorm_name, room_id, suite_uuid, max_occupancy, current_occupancy) VALUES ({dorm}, '{dorm_name}', '{room_number}', '{suite_uuid}', {room_type}, 0) RETURNING room_uuid;"
        room_uuids.append(connection.execute(text(query)).fetchone()[0])
        # connection.commit()
        actual_starting += room_offset
    
    room_uuid_string = ""
    for room_uuid in room_uuids:
        room_uuid_string += f"'{room_uuid}'::UUID, "
    room_uuid_string = room_uuid_string[:-2]
    # update the suite which contains the room
    query = f"UPDATE Suites SET rooms =ARRAY[{room_uuid_string}] WHERE suite_uuid = '{suite_uuid}';"
    connection.execute(text(query))
    # connection.commit()

def create_normal_inner_dorm_floor_list(single_list: List[int]) -> List[dict]:
    # create the list of suites for the floor
    floor_list = []

    # create wings:
    for i in range(1, 15, 2):
        if i in single_list:
            floor_list.append({
                "order": [1, 1],
                "starting_room_number_without_floor": i,
                "alternative": False,
                "offset": 2
            })
        else:
            floor_list.append({
                "order": [2, 2],
                "starting_room_number_without_floor": i,
                "alternative": True,
                "offset": 2
            })
    
    # create the middle:
    # each key which is the room base is mapped to the offset of the next room
    middle = {
        19: 4,
        20: 4,
        21: 4,
        22: 4,
        27: 1
    }
    for key, value in middle.items():
        if key in single_list:
            floor_list.append({
                "order": [1, 1],
                "starting_room_number_without_floor": key,
                "alternative": False,
                "offset": value
            })
        else:
            floor_list.append({
                "order": [2, 2],
                "starting_room_number_without_floor": key,
                "alternative": True,
                "offset": value
            })

    return floor_list

def populate_east():
    # start a transaction
    with engine.connect() as connection:
        query = "DELETE FROM Rooms WHERE dorm = 1; DELETE FROM Suites WHERE dorm = 1; "
        result = connection.execute(text(query))
        # connection.commit()
        east_single_list = [19, 20, 22, 27]
        east_floor = create_normal_inner_dorm_floor_list(east_single_list)
        for floor in range(2):
            for suite in east_floor:
                create_normal_inner_suite(suite["starting_room_number_without_floor"], suite["order"], floor, suite["offset"], 1, suite["alternative"], connection)
        connection.commit()

def populate_north():
    # start a transaction
    with engine.connect() as connection:
        query = "DELETE FROM Rooms WHERE dorm = 2; DELETE FROM Suites WHERE dorm = 2; "
        result = connection.execute(text(query))
        # connection.commit()
        north_single_list = [19, 20, 22, 27]
        north_floor = create_normal_inner_dorm_floor_list(north_single_list)
        for floor in range(2):
            for suite in north_floor:
                create_normal_inner_suite(suite["starting_room_number_without_floor"], suite["order"], floor, suite["offset"], 2, suite["alternative"], connection)
        connection.commit()

def populate_west():
    # start a transaction
    with engine.connect() as connection:
        query = "DELETE FROM Rooms WHERE dorm = 4; DELETE FROM Suites WHERE dorm = 4; "
        result = connection.execute(text(query))
        # connection.commit()
        west_single_list = [19, 20, 21, 27]
        west_floor = create_normal_inner_dorm_floor_list(west_single_list)
        for floor in range(2):
            for suite in west_floor:
                create_normal_inner_suite(suite["starting_room_number_without_floor"], suite["order"], floor, suite["offset"], 4, suite["alternative"], connection)
        connection.commit()
        
        
# Atwood has three different types of configurations: 
# 1. a suite with 4 singles, 1 double, and 1 triple
# 2. a suite with 2 doubles
# 3. a double room
        
def create_atwood_suite(starting_room_number_without_floor: int, order: List[int], floor: int, alternative: bool, connection: object):
    # start a transaction
    # with engine.connect() as connection:
    # create a suite in atwood
    query = f"INSERT INTO Suites (dorm, dorm_name, room_count, floor, alternative_pull) VALUES (5, 'Atwood', {len(order)}, {floor}, {alternative}) RETURNING suite_uuid;"
    result = connection.execute(text(query))
    suite_uuid = result.fetchone()[0]
    # connection.commit()

    room_uuids = []

    # insert the rooms into the rooms table
    actual_starting = starting_room_number_without_floor + ((floor+1) * 100)
    for room_type in order:
        # get the room number
        room_number = actual_starting
        # insert the room into the rooms table
        query = f"INSERT INTO Rooms (dorm, dorm_name, room_id, suite_uuid, max_occupancy, current_occupancy, lock_pulled) VALUES (5, 'Atwood', '{room_number}', '{suite_uuid}', {room_type}, 0, false) RETURNING room_uuid;"
        room_uuids.append(connection.execute(text(query)).fetchone()[0])
        # connection.commit()
        actual_starting += 2
    
    room_uuid_string = ""
    for room_uuid in room_uuids:
        room_uuid_string += f"'{room_uuid}'::UUID, "
    room_uuid_string = room_uuid_string[:-2]
    # update the suite which contains the room
    query = f"UPDATE Suites SET rooms = ARRAY[{room_uuid_string}] WHERE suite_uuid = '{suite_uuid}';"
    connection.execute(text(query))
    # connection.commit()
        
def populate_atwood():
    with engine.connect() as connection:
        query = "DELETE FROM Rooms WHERE dorm = 5; DELETE FROM Suites WHERE dorm = 5;"
        result = connection.execute(text(query))
        # connection.commit()
        atwood_floor = [
            {
                "order" : [2, 1, 1, 3, 1, 1],
                "starting_room_number_without_floor" : 1,
                "alternative" : False
            },
            {
                "order" : [2, 2],
                "starting_room_number_without_floor" : 13,
                "alternative" : True
            },
            {
                "order" : [2, 1, 1, 3, 1, 1],
                "starting_room_number_without_floor" : 0,
                "alternative" : False
            },
            {
                "order" : [2],
                "starting_room_number_without_floor" : 12,
                "alternative" : False
            },
            {
                "order" : [2],
                "starting_room_number_without_floor" : 14,
                "alternative" : False
            },
            {
                "order" : [1, 1, 3, 1, 1, 2],
                "starting_room_number_without_floor" : 16,
                "alternative" : False
            },
        ] 

        for floor in range(3):
            for suite in atwood_floor:
                # no column double on 3rd floor
                if floor == 2 and suite["alternative"]:
                    continue
                create_atwood_suite(suite["starting_room_number_without_floor"], suite["order"], floor, suite["alternative"], connection)
        connection.commit()

'''
sample json format:
{
    "floors": [
        {
            "suites": [
                {
                    "rooms": [
                        {
                            "room_number": "101A",
                            "capacity": 1
                        },
                        {
                            "room_number": "101B",
                            "capacity": 1
                        },
                        {
                            "room_number": "101C",
                            "capacity": 1
                        },
                        {
                            "room_number": "101D",
                            "capacity": 2
                        }
                    ],
                    "alternative_pull": false
                },
                {
                    "rooms": [
                        {
                            "room_number": "102A",
                            "capacity": 1
                        },
                        {
                            "room_number": "102B",
                            "capacity": 1
                        },
                        {
                            "room_number": "102C",
                            "capacity": 1
                        },
                        {
                            "room_number": "102D",
                            "capacity": 2
                        }
                    ],
                    "alternative_pull": false
                },
                {
                    "rooms": [
                        {
                            "room_number": "103A",
                            "capacity": 1
                        },
                        {
                            "room_number": "103B",
                            "capacity": 1
                        },
                        {
                            "room_number": "103C",
                            "capacity": 1
                        },
                        {
                            "room_number": "103D",
                            "capacity": 2
                        }
                    ],
                    "alternative_pull": false
                },
                {
                    "rooms": [
                        {
                            "room_number": "105A",
                            "capacity": 1
                        },
                        {
                            "room_number": "105B",
                            "capacity": 1
                        },
                        {
                            "room_number": "105D",
                            "capacity": 2
                        }
                    ],
                    "alternative_pull": false
                },
                {
                    "rooms": [
                        {
                            "room_number": "106A",
                            "capacity": 1
                        },
                        {
                            "room_number": "106B",
                            "capacity": 1
                        },
                        {
                            "room_number": "106C",
                            "capacity": 1
                        },
                        {
                            "room_number": "106D",
                            "capacity": 2
                        }
                    ],
                    "alternative_pull": false
                },
                {
                    "rooms": [
                        {
                            "room_number": "107A",
                            "capacity": 1
                        },
                        {
                            "room_number": "107B",
                            "capacity": 1
                        },
                        {
                            "room_number": "107C",
                            "capacity": 1
                        },
                        {
                            "room_number": "107D",
                            "capacity": 2
                        }
                    ],
                    "alternative_pull": false
                },
                {
                    "rooms": [
                        {
                            "room_number": "108A",
                            "capacity": 1
                        },
                        {
                            "room_number": "108B",
                            "capacity": 1
                        },
                        {
                            "room_number": "108C",
                            "capacity": 1
                        },
                        {
                            "room_number": "108D",
                            "capacity": 2
                        }
                    ],
                    "alternative_pull": false
                }
            ]
        },
        {
            "suites": [
                {
                    "rooms": [
                        {
                            "room_number": "201A",
                            "capacity": 1
                        },
                        {
                            "room_number": "201B",
                            "capacity": 1
                        },
                        {
                            "room_number": "201C",
                            "capacity": 1
                        },
                        {
                            "room_number": "201D",
                            "capacity": 2
                        }
                    ],
                    "alternative_pull": false
                },
                {
                    "rooms": [
                        {
                            "room_number": "202A",
                            "capacity": 1
                        },
                        {
                            "room_number": "202B",
                            "capacity": 1
                        },
                        {
                            "room_number": "202C",
                            "capacity": 1
                        },
                        {
                            "room_number": "202D",
                            "capacity": 2
                        }
                    ],
                    "alternative_pull": false
                },
                {
                    "rooms": [
                        {
                            "room_number": "203A",
                            "capacity": 1
                        },
                        {
                            "room_number": "203B",
                            "capacity": 1
                        },
                        {
                            "room_number": "203C",
                            "capacity": 1
                        },
                        {
                            "room_number": "203D",
                            "capacity": 2
                        }
                    ],
                    "alternative_pull": false
                },
                {
                    "rooms": [
                        {
                            "room_number": "204A",
                            "capacity": 1
                        },
                        {
                            "room_number": "204B",
                            "capacity": 1
                        },
                        {
                            "room_number": "204D",
                            "capacity": 2
                        }
                    ],
                    "alternative_pull": false
                },
                {
                    "rooms": [
                        {
                            "room_number": "205A",
                            "capacity": 1
                        },
                        {
                            "room_number": "205B",
                            "capacity": 1
                        },
                        {
                            "room_number": "205D",
                            "capacity": 2
                        }
                    ],
                    "alternative_pull": false
                },
                {
                    "rooms": [
                        {
                            "room_number": "206A",
                            "capacity": 1
                        },
                        {
                            "room_number": "206B",
                            "capacity": 1
                        },
                        {
                            "room_number": "206C",
                            "capacity": 1
                        },
                        {
                            "room_number": "206D",
                            "capacity": 2
                        }
                    ],
                    "alternative_pull": false
                },
                {
                    "rooms": [
                        {
                            "room_number": "207A",
                            "capacity": 1
                        },
                        {
                            "room_number": "207B",
                            "capacity": 1
                        },
                        {
                            "room_number": "207C",
                            "capacity": 1
                        },
                        {
                            "room_number": "207D",
                            "capacity": 2
                        }
                    ],
                    "alternative_pull": false
                },
                {
                    "rooms": [
                        {
                            "room_number": "208A",
                            "capacity": 1
                        },
                        {
                            "room_number": "208B",
                            "capacity": 1
                        },
                        {
                            "room_number": "208C",
                            "capacity": 1
                        },
                        {
                            "room_number": "208D",
                            "capacity": 2
                        }
                    ],
                    "alternative_pull": false
                },
                {
                    "rooms": [
                        {
                            "room_number": "209A",
                            "capacity": 1
                        },
                        {
                            "room_number": "209D",
                            "capacity": 2
                        }
                    ],
                    "alternative_pull": false
                },
                {
                    "rooms": [
                        {
                            "room_number": "210A",
                            "capacity": 1
                        },
                        {
                            "room_number": "210D",
                            "capacity": 2
                        }
                    ],
                    "alternative_pull": false
                }
            ]
        }
    ]
}
'''
def populate_using_json(dorm_id: int, dorm_name: str, json_file: str):
    import json
    with open(json_file, 'r') as file:
        data = json.load(file)
        with engine.connect() as connection:
            query = f"DELETE FROM Rooms WHERE dorm = {dorm_id}; DELETE FROM Suites WHERE dorm = {dorm_id};"
            result = connection.execute(text(query))
            # connection.commit()
            for floor in range(len(data["floors"])):
                for suite in data["floors"][floor]["suites"]:
                    query = f"INSERT INTO Suites (dorm, dorm_name, room_count, floor, alternative_pull) VALUES ({dorm_id}, '{dorm_name}', {len(suite['rooms'])}, {floor}, {suite['alternative_pull']}) RETURNING suite_uuid;"
                    result = connection.execute(text(query))
                    suite_uuid = result.fetchone()[0]
                    # connection.commit()

                    room_uuids = []

                    # insert the rooms into the rooms table
                    for room in suite["rooms"]:
                        # insert the room into the rooms table
                        query = f"INSERT INTO Rooms (dorm, dorm_name, room_id, suite_uuid, max_occupancy, current_occupancy) VALUES ({dorm_id}, '{dorm_name}', '{room['room_number']}', '{suite_uuid}', {room['capacity']}, 0) RETURNING room_uuid;"
                        room_uuids.append(connection.execute(text(query)).fetchone()[0])
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
    populate_atwood()
    populate_east()
    populate_north()
    populate_west()
    populate_using_json(3, "South", "south.json")
    populate_using_json(6, "Sontag", "sontag.json")
    populate_using_json(7, "Case", "case.json")
    populate_using_json(8, "Drinkward", "drinkward.json")
    populate_using_json(9, "Linde", "linde.json")


populate_all()