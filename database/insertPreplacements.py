# read in preplacements.csv into a dataframe
import os
import requests
from dotenv import load_dotenv
from sqlalchemy.sql import text
from sqlalchemy import create_engine
import pandas as pd

df = pd.read_csv('preplacements.csv')

# print the first 5 rows of the dataframe
print(df.head())

# Dorm Mapping
# 1 = East
# 2 = North
# 3 = South
# 4 = West
# 5 = Atwood
# 6 = Sontag
# 7 = Case
# 8 = Drinkward
# 9 = Linde


# import env variables

# the CSV columns should be:
# ID, First Name, Last Name, Email, Rising Class Yr (next year), Planned Grad Sess/Year, Dorm, Room, Preplacement Reason


dotenv_path = os.path.join(os.getcwd(), '.env')
print(dotenv_path)

load_dotenv(dotenv_path=dotenv_path, verbose=True)

sql_pass = os.environ.get('SQL_PASS')
sql_ip = os.environ.get('SQL_IP')
sql_db_name = os.environ.get('SQL_DB_NAME')
sql_user = os.environ.get('SQL_USER')

CONNSTR = f'postgresql://{sql_user}:{sql_pass}@{sql_ip}/{sql_db_name}'

engine = create_engine(CONNSTR)

# print(CONNSTR)

with engine.connect() as connection:
    # loop through the dataframe and insert each row into the database
    for index, row in df.iterrows():
        # create a dictionary to store the values for each row
        values = {
            'first_name': row['First Name'],
            'last_name': row['Last Name'],
            'email': row['Email'],
            'reslife_role': row['Preplacement Reason']
        }
        # if values['reslife_role'] does not contain the word 'Proctor' or 'Mentor' then set it to 'None'
        print(values['reslife_role'])
        if str(values['reslife_role']) == "nan" or ('proctor' not in values['reslife_role'].lower() and 'mentor' not in values['reslife_role'].lower()):
            values['reslife_role'] = 'none'

        values['reslife_role'] = values['reslife_role'].lower().strip()
        year = 0
        in_dorm = 0
        draw_number = 0
        preplaced = True

        # create a query to insert the values into the database
        query = f"INSERT INTO users (first_name, last_name, draw_number, year, preplaced, in_dorm, email, reslife_role) VALUES ('{values['first_name']}', '{values['last_name']}', {draw_number}, {year}, {preplaced}, {in_dorm}, '{values['email']}', '{values['reslife_role']}')"
        # execute the query
        result = connection.execute(text(query))

        print(
            f'Inserted {values["first_name"]} {values["last_name"]} into the database')

    connection.commit()


with engine.connect() as connection:
    # query the database to get all the users ids which were assigned by the database and insert them into the dataframe
    query = "SELECT id, first_name, last_name FROM users"
    result = connection.execute(text(query))
    # create a dictionary to store the ids of the users
    user_ids = {}
    # loop through the result and add the id to the dictionary
    for row in result:
        user_ids[f'{row[1]} {row[2]}'] = row[0]

    # loop through the dataframe and add the user ids to the dataframe
    for index, row in df.iterrows():
        df.at[index,
              'id'] = user_ids[f'{row["First Name"]} {row["Last Name"]}']

    connection.commit()

jwt = "INSERT_JWT_HERE"

# link is localhost:8080/rooms/preplace/:roomuuid
# query the database to get all the rooms
with engine.connect() as connection:
    query = "SELECT room_uuid, dorm_name, room_id FROM rooms"
    result = connection.execute(text(query))
    # create a dictionary to store the room ids
    room_ids = {}
    # loop through the result and add the room ids to the dictionary
    for row in result:
        room_ids[f'{row[1]} {row[2]}'] = row[0]

    connection.commit()

# find all rows where the which share a common value for both Dorm and Room
# iterate through the groups that share a common value for both Dorm and Room and print the group
for name, group in df.groupby(['Dorm', 'Room']):
    # find the room_uuid for the group
    room_uuid = room_ids[f'{name[0]} {name[1]}']
    print(f'Room UUID: {room_uuid}')

    # create a rest call to the server to preplace the users in the group
    # rest request body contains json with:
    # type PreplacedRequest struct {
    # , ProposedOccupants []int  `json:"proposedOccupants"`
    # , UserJWT           string `json:"userJWT"`
    # }
    # use the id of the users in the group to create the proposed occupants list
    users_in_group = group['id'].tolist()
    # convert to integers
    users_in_group = [int(user) for user in users_in_group]

    print(f'Users in group: {users_in_group}')
    proposed_occupants = []
    for user in users_in_group:
        proposed_occupants.append(user)
    # create the json body for the rest request
    json_body = {
        'proposedOccupants': proposed_occupants,
        'userJWT': jwt
    }

    # create the headers for the rest request
    headers = {
        'Content-Type': 'application/json',
        'Authorization': f'Bearer {jwt}'
    }

    # create the url for the rest request
    url = f'http://localhost:8080/rooms/preplace/{room_uuid}'

    print(url)

    response = requests.post(url, headers=headers, json=json_body, timeout=10)

    # print the response
    # print(response.json())
