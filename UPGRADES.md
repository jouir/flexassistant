# General notes

In the next commands, you should replace `flexassistant.db` by the content of the `database_file` setting if it's has been changed.

# 1.2 to 1.3

The balance has reached maximum of **int64** type so the column should be converted to **real** to handle large numbers:

```
sqlite3 flexassistant.db < migrations/1.2_to_1.3.sql
```

# 1.0 to 1.1

Some numeric types have been updated from **float64** to **int64**. Upgrade the database types by running the following migration:

```
sqlite3 flexassistant.db < migrations/1.0_to_1.1.sql
```