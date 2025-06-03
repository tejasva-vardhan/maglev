-- name: GetAgency :one
SELECT
    *
FROM
    agencies
WHERE
    id = ?
LIMIT
    1;

-- name: ListAgencies :many
SELECT
    *
FROM
    agencies
ORDER BY
    id;

-- name: CreateAgency :one
INSERT
OR REPLACE INTO agencies (
    id,
    name,
    url,
    timezone,
    lang,
    phone,
    fare_url,
    email
)
VALUES
    (?, ?, ?, ?, ?, ?, ?, ?) RETURNING *;

-- name: CreateRoute :one
INSERT
OR REPLACE INTO routes (
    id,
    agency_id,
    short_name,
    long_name,
    desc,
    type,
    url,
    color,
    text_color,
    continuous_pickup,
    continuous_drop_off
)
VALUES
    (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) RETURNING *;

-- name: CreateStop :one
INSERT
OR REPLACE INTO stops (
    id,
    code,
    name,
    desc,
    lat,
    lon,
    zone_id,
    url,
    location_type,
    timezone,
    wheelchair_boarding,
    platform_code
)
VALUES
    (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) RETURNING *;

-- name: CreateCalendar :one
INSERT
OR REPLACE INTO calendar (
    id,
    monday,
    tuesday,
    wednesday,
    thursday,
    friday,
    saturday,
    sunday,
    start_date,
    end_date
)
VALUES
    (?, ?, ?, ?, ?, ?, ?, ?, ?, ?) RETURNING *;

-- name: CreateShape :one
INSERT
OR REPLACE INTO shapes (shape_id, lat, lon, shape_pt_sequence)
VALUES
    (?, ?, ?, ?) RETURNING *;

-- name: CreateStopTime :one
INSERT
OR REPLACE INTO stop_times (
    trip_id,
    arrival_time,
    departure_time,
    stop_id,
    stop_sequence,
    stop_headsign,
    pickup_type,
    drop_off_type,
    timepoint
)
VALUES
    (?, ?, ?, ?, ?, ?, ?, ?, ?) RETURNING *;

-- name: CreateTrip :one
INSERT
OR REPLACE INTO trips (
    id,
    route_id,
    service_id,
    trip_headsign,
    trip_short_name,
    direction_id,
    block_id,
    shape_id,
    wheelchair_accessible,
    bikes_allowed
)
VALUES
    (?, ?, ?, ?, ?, ?, ?, ?, ?, ?) RETURNING *;

-- name: ListRoutes :many
SELECT
    id,
    agency_id,
    short_name,
    long_name,
    "desc",
    type,
    url,
    color,
    text_color,
    continuous_pickup,
    continuous_drop_off
FROM
    routes
ORDER BY
    agency_id,
    id;

-- name: GetRouteIDsForStop :many
SELECT DISTINCT
    routes.agency_id || '_' || routes.id AS route_id
FROM
    stop_times
    JOIN trips ON stop_times.trip_id = trips.id
    JOIN routes ON trips.route_id = routes.id
WHERE
    stop_times.stop_id = ?;

-- name: GetAgencyForStop :one
SELECT DISTINCT
    a.id,
    a.name,
    a.url,
    a.timezone,
    a.lang,
    a.phone,
    a.fare_url,
    a.email
FROM
    agencies a
    JOIN routes r ON a.id = r.agency_id
    JOIN trips t ON r.id = t.route_id
    JOIN stop_times st ON t.id = st.trip_id
WHERE
    st.stop_id = ?
ORDER BY
    a.id
LIMIT
    1;

-- name: GetStopIDsForAgency :many
SELECT
    s.id
FROM
    stops s;

-- name: GetRouteIDsForAgency :many
SELECT
    r.id
FROM
    routes r
    JOIN agencies a ON r.agency_id = a.id
WHERE
    a.id = ?;
