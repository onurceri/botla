Whenever you make a change in the backend project, use Makefile to run linter, vulnerability scanner, formatter, vet, shadow checker etc and make sure that there is no error and the code is clean and ready to be merged.

```bash
make fmt
make imports
make lint
make vet
make shadow
make vuln
```

If you need to check data from database, you can use the following command to run the database container:

```bash
make db
```

While writing backend code, make sure that you are not shadowing any variable.
Be aware of this error and avoid it: 

```bash
shadow: declaration of "err" shadows declaration at ..

To run the backend project, you can use Docker to connect to the database container and run any select command.

```bash
docker exec botla-postgres psql -U botla -d botla_dev -c {COMMAND}
```