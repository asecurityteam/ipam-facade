-- from the higher-voted answer here:  https://stackoverflow.com/questions/3327312/how-can-i-drop-all-the-tables-in-a-postgresql-database
DROP SCHEMA public CASCADE;
CREATE SCHEMA
public;
GRANT ALL ON SCHEMA public TO public;