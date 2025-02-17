CREATE TABLE all_datatypes (
    tinyint_col TINYINT, 
    smallint_col SMALLINT,
    mediumint_col MEDIUMINT,
    int_col INT,
    bigint_col BIGINT,
    float_col FLOAT,
    double_col DOUBLE,
    decimal_col DECIMAL(10,2),

    char_col CHAR(10),
    varchar_col VARCHAR(255),
    tinytext_col TINYTEXT,
    text_col TEXT,
    mediumtext_col MEDIUMTEXT,
    longtext_col LONGTEXT,

	binary_col BINARY(10),      
    varbinary_col VARBINARY(255),
    tinyblob_col TINYBLOB,
    blob_col BLOB,
    mediumblob_col MEDIUMBLOB,
    longblob_col LONGBLOB,

    enum_col ENUM('small', 'medium', 'large'),
    set_col SET('a', 'b', 'c', 'd'),

    -- Date & Time Data Types
    date_col DATE,
    datetime_col DATETIME,
    timestamp_col TIMESTAMP,
    time_col TIME,
    year_col YEAR,

    -- Special Data Types
    boolean_col BOOLEAN,
    bit_col BIT(8),
    json_col JSON,

    -- Spatial Data Types (For GIS/Geometric Data)
    geometry_col GEOMETRY,
    -- point_col POINT, clickhouse is throwing error for this column type (clickhouse Nested type Point cannot be inside Nullable type)
    linestring_col LINESTRING,
    polygon_col POLYGON,
    multipoint_col MULTIPOINT,
    multilinestring_col MULTILINESTRING,
    multipolygon_col MULTIPOLYGON,
    geometrycollection_col GEOMETRYCOLLECTION
);


INSERT INTO all_datatypes (
    tinyint_col, smallint_col, mediumint_col, int_col, bigint_col,
    float_col, double_col, decimal_col,
    char_col, varchar_col, tinytext_col, text_col, mediumtext_col, longtext_col,
    binary_col, varbinary_col, tinyblob_col, blob_col, mediumblob_col, longblob_col,
    enum_col, set_col,
    date_col, datetime_col, timestamp_col, time_col, year_col,
    boolean_col, bit_col, json_col,
    geometry_col,
	-- point_col, 
	linestring_col, polygon_col,
    multipoint_col, multilinestring_col, multipolygon_col, geometrycollection_col
) VALUES 
(
    -- Row 1: All NULL values
    NULL, NULL, NULL, NULL, NULL, 
    NULL, NULL, NULL,
    NULL, NULL, NULL, NULL, NULL, NULL,
    NULL, NULL, NULL, NULL, NULL, NULL,
    NULL, NULL,
    NULL, NULL, NULL, NULL, NULL,
    NULL, NULL, NULL,
    NULL, 
	-- NULL, 
	NULL, NULL,
    NULL, NULL, NULL, NULL
), 
(
    -- Row 2: Non NUll values
    127, 32767, 8388607, 2147483647, 999999999999999999,
    1.1, 2.2, 3.3,
    'C', 'VarChar', 'Tiny Text', 'Text', 'Medium Text', 'Long text content',
    'Binary', 'VarBinary', 'Tiny Blob', 'Blob', 'Medium Blob', 'Long Blob',
    'medium', 'a,b',
    '2024-02-14', '2025-02-14 12:34:56', '2025-02-14 12:34:56', '12:34:56', 2024,
    1, '1', '{"key": "value"}',
    ST_GeomFromText('POINT(1 1)'), 
	-- ST_GeomFromText('POINT(2 2)'),
    ST_GeomFromText('LINESTRING(0 0, 1 1)'),
    ST_GeomFromText('POLYGON((0 0, 1 1, 1 0, 0 0))'),
    ST_GeomFromText('MULTIPOINT((1 1), (2 2))'),
    ST_GeomFromText('MULTILINESTRING((0 0, 1 1), (1 1, 2 2))'),
    ST_GeomFromText('MULTIPOLYGON(((0 0, 1 1, 1 0, 0 0)))'),
    ST_GeomFromText('GEOMETRYCOLLECTION(POINT(1 1), LINESTRING(0 0, 1 1))')
), 
(
    -- Row 3: All Zero Values
    0, 0, 0, 0, 0,
    0.0, 0.0, 0.00,
    '', '', '', '', '', '',
    '', '', '', '', '', '',
    'small', '',
    '1970-01-01', '1970-01-01 00:00:00', NULL, '00:00:00', 1970,
    0, B'0', '{}',
    NULL, 
	-- NULL, 
	NULL, NULL,
    NULL, NULL, NULL, NULL
);
