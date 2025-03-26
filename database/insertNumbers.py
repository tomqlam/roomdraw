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

import os
from dotenv import load_dotenv
from sqlalchemy.sql import text
from sqlalchemy import create_engine
import pandas as pd
# the CSV columns should be:
# First Name, Last Name, Year, In Dorm, Number, Email

# read file numbers.csv
numbers = pd.read_csv('numbers.csv')
# using the mapping for dorm, convert the "In Dorm" column to the corresponding dorm id there are also NaN values
dorm_mapping = {
    'East': 1,
    'North': 2,
    'South': 3,
    'West': 4,
    'Atwood': 5,
    'Sontag': 6,
    'Case': 7,
    'Drinkward': 8,
    'Linde': 9
}

numbers['In Dorm'] = numbers['In Dorm'].map(dorm_mapping)
# fill the NaN values with 0
numbers['In Dorm'] = numbers['In Dorm'].fillna(0)
# convert the In Dorm column to integer
numbers['In Dorm'] = numbers['In Dorm'].astype(int)
# print the dataframe

# SR = 4, JR = 3, SO = 2, FR = 1
year_mapping = {
    'SR': 'senior',
    'JR': 'junior',
    'SO': 'sophomore',
    'FR': 'freshman'
}

numbers['Year'] = numbers['Year'].map(year_mapping)
# convert the Year column to integer

print(numbers)

# import env variables

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
    for index, row in numbers.iterrows():
        # create a dictionary to store the values for each row
        values = {
            'first_name': row['First Name'],
            'last_name': row['Last Name'],
            'email': row['Email'],
            'year': row['Year'],
            'in_dorm': row['In Dorm'],
            'draw_number': row['Number']
        }
        # for names escape out the single quotes
        values['first_name'] = values['first_name'].replace("'", "''")
        values['last_name'] = values['last_name'].replace("'", "''")
        # insert the values into the database
        query = f"INSERT INTO users (first_name, last_name, draw_number, year, preplaced, in_dorm, email) VALUES ('{values['first_name']}', '{values['last_name']}', {values['draw_number']}, '{values['year']}', {False}, {values['in_dorm']}, '{values['email']}')"
        result = connection.execute(text(query))

    connection.commit()
print('done')
