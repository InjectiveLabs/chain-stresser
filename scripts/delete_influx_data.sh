#!/bin/bash

# InfluxDB connection settings
INFLUX_HOST="localhost:8086"
DATABASE="your_database_name"

# Time range (nanosecond precision)
START_TIME=1749016840234739465
END_TIME=1758016214377204687

# Delete data from all measurements in the time range
influx -host "$INFLUX_HOST" -database "$DATABASE" -execute "DELETE WHERE time >= ${START_TIME} AND time <= ${END_TIME}"

echo "Deleted data between $START_TIME and $END_TIME from all measurements"
