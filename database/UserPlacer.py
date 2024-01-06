import random
from sqlalchemy import Boolean, Column, Integer, String, create_engine
from sqlalchemy.sql import text

# import env variables
import os
import requests
from dotenv import load_dotenv

users = []
results = 400
# get env variable called neon_pass from env file
dotenv_path = os.path.join(os.path.dirname(__file__), '.env')
print(dotenv_path)

load_dotenv(dotenv_path=dotenv_path, verbose=True)

cloud_sql_pass = os.environ.get('CLOUD_SQL_PASS')
cloud_sql_ip = os.environ.get('CLOUD_SQL_IP')

CONNSTR = f'postgresql://postgres:{cloud_sql_pass}@{cloud_sql_ip}/test-db'

engine = create_engine(CONNSTR)

with engine.connect() as connection:
    seniors = results//3
    juniors = results*2 // 3 - results // 3
    sophomores = results - results*2 // 3

    # pull all users into the users array
    query = "SELECT * FROM Users;"
    result = connection.execute(text(query))
    users = result.fetchall()
    # connection.commit()

    # pull all rooms into the rooms array
    query = "SELECT * FROM Rooms;"
    result = connection.execute(text(query))
    rooms = result.fetchall()
    # connection.commit()
    used = {}
        
    # delete all value from column occupants
    query = "UPDATE Rooms SET occupants = '{}'; UPDATE Rooms SET current_occupancy = 0"
    result = connection.execute(text(query))

    # reset user table (remove room_uuid and preplaced)
    query = "UPDATE Users SET room_uuid = NULL;"
    result = connection.execute(text(query))

    # connection.commit()

    # reset all preplacements
    # query = "UPDATE Users SET preplaced = false;"
    # result = connection.execute(text(query))
    # connection.commit()
    # the room database has a column called occupants which is of type array of uuids
    usedRooms = {}
    for i in range(200):
        # select a room at random
        room = random.choice(rooms)
        while (room.room_id in usedRooms):
            room = random.choice(rooms)
        usedRooms[room.room_id] = True

        max_occupancy = room.max_occupancy
        current_occupancy = room.current_occupancy
        if(current_occupancy != 0):
            continue
        occupants = room.occupants
        if(room.occupants == None):
            occupants = []
        # select max_occupancy - current_occupancy users at random, mark them as used
        # and add them to the occupants array
        print(max_occupancy)
        print(current_occupancy)
        for j in range(max_occupancy-current_occupancy):
            user = random.choice(users)
            while (user.id in used):
                user = random.choice(users)
            used[user.id] = True
            print(user.id)
            occupants.append(user.id)
            query = f"UPDATE Users SET room_uuid = \'{str(room.room_uuid)}\' WHERE id=\'{str(user.id)}\';"
            result = connection.execute(text(query))
        current_occupancy = len(occupants)
        # update the room with the new occupants array
        formatted = str(occupants).replace("\'", "\"").replace("[","'{").replace("]","}'")
        print(formatted)
        # query = f"UPDATE Rooms SET occupants = {formatted} WHERE room_id=\'{str(room.room_id)}\';"
        # also set current_occupancy to the new value
        query = f"UPDATE Rooms SET occupants = {formatted}, current_occupancy = {current_occupancy} WHERE room_uuid=\'{str(room.room_uuid)}\';"
        result = connection.execute(text(query))
        print(query)
    connection.commit()
    