#### For running docker container ***`schema-generator`*** with generator util:
```shell
make docker-run in=$HOME/work/latest-ccs-installer out=/var -e OVERRIDE_SCHEMA_PATH=/schemas/values.schema.json -e VALUES_PATH=/var/values.yaml -e SCHEMA_PATH=/home/values.schema.json
```
##### where:
- ***`in`*** - using for volume mounting directory with ***`Helm`*** chart into container file system (set mount from host)
- ***`out`*** - using for volume mounting directory with ***`Helm`*** chart into container file system (set mount to container)
##### environments:
- ***`OVERRIDE_SCHEMA_PATH`*** - path of json file with patch schema for merging. By default, override schema file is located by path: ****`/schemas/values.schema.json`**** in container. (Consider where you mounted directories)
- ***`VALUES_PATH`*** - path of current chart ****`values.yaml`**** file (Consider where you mounted directories)
- ***`SCHEMA_PATH`*** - file path where to write the json schema. By default, write into ****`/tmp`**** directory in workspace (Consider where you mounted directories)

#### For building docker image with generator util:
```shell
make docker-build
```

To generate a json schema based on an existing helm chart located in the path:
****``****
relative to the root repository, simply run the command: 
```shell
make generate
```
This will generate a ****`values.schema.json`**** file in the chart folder based on the chart's ****`values.yaml`****.
For run make command from any another directory, try:
```shell
make -C <path to repository>/json_schema_converter generate
```