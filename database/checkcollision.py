import pandas as pd

# read file numbers.csv
numbers = pd.read_csv('numbers.csv')

# read file preplacements.csv
preplacements = pd.read_csv('preplacements.csv')

# using the email column see if there are any users in both numbers and preferences
# if there are, print out the email

collisions = pd.merge(numbers, preplacements, on='Email', how='inner')

print(collisions)

# also add rows where both the first and last name match
collisions2 = pd.merge(numbers, preplacements, on=[
                       'First Name', 'Last Name'], how='inner')

print(collisions2)

# write the collisions to a file
collisions.to_csv('collisions.csv', index=False)
