import random
from sqlalchemy import create_engine
from sqlalchemy.sql import text
from uuid import uuid4

# import env variables
import os
import requests
from dotenv import load_dotenv


dotenv_path = os.path.join(os.path.dirname(__file__), '.env')
print(dotenv_path)

load_dotenv(dotenv_path=dotenv_path, verbose=True)

cloud_sql_pass = os.environ.get('CLOUD_SQL_PASS')
cloud_sql_ip = os.environ.get('CLOUD_SQL_IP')
cloud_sql_db_name = os.environ.get('CLOUD_SQL_DB_NAME')
cloud_sql_user = os.environ.get('CLOUD_SQL_USER')

CONNSTR = f'postgresql://{cloud_sql_user}:{cloud_sql_pass}@{cloud_sql_ip}/{cloud_sql_db_name}'

engine = create_engine(CONNSTR)

# i need to query the api https://randomuser.me/api/?inc=name with a get request

users = []
results = 700

response = requests.get('https://randomuser.me/api/?nat=au,br,ca,ch,de,dk,es,fi,fr,gb,ie,in,mx,nl,no,nz,rs,tr,ua,us&inc=name&results='+str(results))
data = response.json()

with engine.connect() as connection:
    # clear the table
    query = "DELETE FROM Users;"
    result = connection.execute(text(query))
    # connection.commit()

    # make a list of number from 1-52 and randomise it
    
    
    seniors = results//3
    juniors = results*2 // 3 - results // 3
    sophomores = results - results*2 // 3
    print(seniors, juniors, sophomores)

    senior_draw_list = list(range(1, seniors+1))
    junion_draw_list = list(range(1, juniors+1))
    sophomore_draw_list = list(range(1, sophomores+1))
    random.shuffle(senior_draw_list)
    random.shuffle(junion_draw_list)
    random.shuffle(sophomore_draw_list)

    for i in range(results):
        first_name = data['results'][i]['name']['first']
        last_name = data['results'][i]['name']['last']

        escape_first_name = first_name.replace("'", "''")
        escape_last_name = last_name.replace("'", "''")
        # random number between 40000000 and 49999999
        # random_id = str(random.randint(40000000, 49999999))
        # students distributed equally between senior, junior, and sophomore
        year = ''
        draw_number = 0
        if (i < seniors):
            year = 'senior'
            draw_number = senior_draw_list.pop()
        elif (i < juniors + seniors):
            year = 'junior'
            draw_number = junion_draw_list.pop()
        else:
            year = 'sophomore'
            draw_number = sophomore_draw_list.pop()
        # 1/10 chance of being a preplaced student
        preplaced = (i%10==0)
        in_dorm = random.randint(1, 9) if (i%2==0 and (not preplaced and year =='senior')) else 0;
        # generate a uuid
        query = f"INSERT INTO Users (first_name, last_name, draw_number, year, preplaced, in_dorm) VALUES ('{first_name}', '{last_name}', {draw_number}, '{year}', {preplaced}, {in_dorm});"
        result = connection.execute(text(query))
    connection.commit()
