# 1.0 to 1.1

Some numeric types have been updated from **float64** to **int64**. Upgrade the database types by running the following migration:

```
sqlite3 flexassistant.db < migrations/1.0_to_1.1.sql
```

And replace `flexassistant.db` by the content of the `database_file` setting if it's has been changed.