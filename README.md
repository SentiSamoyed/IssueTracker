# IssueTracker

The issue tracker service of the project [**_SentiStrength_**](https://github.com/SentiSamoyed/SentiStrength).

## Usage

### Build

```bash
> cd src
> go build -o IssueTracker
```

### Run

Before running, please set up your own `config.yaml`

- `sever.addr` is the address that your server listens to.
- `datasource.user` is the username of your database.
- `datasource.password` is the name of the environment variable that stores your database's password.
- `datasource.suffix` contains the address of your database.

And you're recommended to set up the env `GH_TOKEN` to your own token, which is used to authorize the GitHub client.

Then you can simply run it by:
```bash
> ./IssueTracker config.yaml
# Usage: ./IssueTracker <path/to/config.yaml>
```

> Note: About the SQL DDL of the database, please take reference to the `sql` directory of [SentiStrength](https://github.com/SentiSamoyed/SentiStrength).
