SELECT
    -1::INT1 as tint1,
    -2::INT2 as tint2,
    -4::INT4 as tint4,
    -8::INT8 as tint8,
    1::UTINYINT as tuint1,
    2::USMALLINT as tuint2,
    4::UINTEGER as tuint4,
    8::UBIGINT as tuint8,
    1::HUGEINT as thugeint,
    4::FLOAT4 as tfloat4,
    8::FLOAT8 as tfloat8,
    1::DECIMAL(18,3) as tdecimal,
    TRUE as tbool,
    ['a','b'] as tlist,
    map {'f1' : 1, 'f2': 2} as tmap,
    {'f1' : 1, 'f2': { 'f3': 3 }} as tstruct,
    TIMESTAMP '2023-01-01' as timestamp,
    uuid() as tuuid
