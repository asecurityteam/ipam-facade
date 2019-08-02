CREATE TABLE
IF NOT EXISTS customers
(
    id INTEGER PRIMARY KEY,
    resource_owner TEXT NOT NULL,
    business_unit TEXT NOT NULL
);

CREATE TABLE
IF NOT EXISTS subnets
(
    id INTEGER PRIMARY KEY,
    network CIDR NOT NULL,
    location TEXT NOT NULL,
    customer_id INTEGER NOT NULL,
    FOREIGN KEY (customer_id) REFERENCES customers (id) ON DELETE CASCADE
);

CREATE TABLE
IF NOT EXISTS ips
(
    id SERIAL PRIMARY KEY,
    ip INET NOT NULL,
    subnet_id INTEGER NOT NULL,
    FOREIGN KEY (subnet_id) REFERENCES subnets (id) ON DELETE CASCADE,
    device_id INTEGER
);