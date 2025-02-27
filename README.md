# go gqlgen app

## How to run project?

Simply use

``docker compose up ``


## Database models
 I chose to implement translations database as a single table Word
 with self relation many to many

![first_er_model.png](project_info/first_er_model.png)

Later I further optimized it to two tables:
- Word table - stores unique words
- Translation table - stores translations between words

![second_er_model.png](project_info/second_er_model.png)